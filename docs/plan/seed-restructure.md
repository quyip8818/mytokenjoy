# Seed 数据架构设计

> **状态：已实施**（代码在 `seed/bootstrap/`、`seed/init.go`、`internal/config/bootstrap.go`）

## 核心原则

1. **所有数据库初始化逻辑收归 `seed/`** — `store/postgres` 只管连接和 DDL schema
2. **prod bootstrap 与 demo 假数据物理分离** — 不同子包、不同入口
3. **prod bootstrap 由 config file 驱动** — selfhosted 通过 YAML 定制初始化参数
4. **internal 不 import seed/** — domain、pkg、infra 禁止依赖 seed；bootstrap 可 import internal（单向依赖）
5. **运行时代码不 import seed/** — domain、pkg、infra 禁止依赖 seed

---

## 目录结构

```
seed/
├── bootstrap/                  # 任何环境首次启动都需要的数据
│   ├── bootstrap.go            # ApplyBootstrap() 总入口
│   ├── config.go               # Config struct + LoadConfig(path) + DefaultConfig()
│   ├── currencies.go           # currencies 表写入
│   ├── companies.go            # companies 表写入
│   ├── permissions.go          # permissions 表写入（通过 internal/infra/permission 读 manifest）
│   ├── roles.go                # roles + role_permission_grants 写入 + ReconcilePresetRoles
│   ├── org.go                  # org_nodes（根部门）+ admin member
│   └── models.go               # models 表写入（可选）
│
├── demo/                       # 仅 BOOTSTRAP_MODE=demo
│   ├── contract/               # 固定 ID、消费常量、demo 密码
│   ├── data/                   # 嵌入 JSON（platform_keys, operation_logs, usage_ledger）
│   ├── filler/                 # 成员生成器
│   ├── snapshot/               # 构建完整 demo Snapshot
│   ├── apply/                  # 写入 demo 表数据
│   ├── runtime/                # 充值、usage（需 store.Store）
│   └── loader.go               # ApplyDemo(ctx, store, cfg)
│
└── seed.go                     # Init(ctx, pool, store, cfg) — 唯一对外入口
```

---

## Bootstrap Config File

### 作用

Selfhosted 部署时，运维通过 YAML 配置文件声明初始化参数。启动时读取一次，写入 DB 后不再依赖。

### 位置

`BOOTSTRAP_CONFIG_PATH` 环境变量。未设置时使用内嵌默认值。

### Schema

```yaml
version: 1

company:
  name: "我的公司"                    # 公司显示名（必填）

admin:                                # 可选，为空则启动后通过 invite 流程添加
  name: "管理员"
  email: "admin@example.com"
  password: ""                        # 为空则首次登录时设置

billing:
  currency: "CNY"                     # 默认 CNY
  quota_per_unit: 500000              # 1 CNY = 500000 quota

models: []                            # 初始模型目录，为空则不预置
# - call_type: "deepseek-v4"
#   name: "DeepSeek V4"
#   provider: "deepseek"
#   input_ratio: 1.0
#   output_ratio: 3.0
```

### 设计约束

- **幂等** — 所有 INSERT 使用 `ON CONFLICT DO NOTHING`（基于下文 unique key）
- **只插不更**  — 手动修改过的数据不会被 bootstrap 覆盖或回滚
- **密码安全** — config 明文密码在 apply 时 bcrypt hash 后写入

### 幂等 key

| 表 | Unique constraint |
|----|-------------------|
| currencies | `(currency)` |
| companies | `(id)` |
| permissions | `(id)` |
| roles | `(company_id, name)` — preset role 按 name 去重 |
| role_permission_grants | `(company_id, role_id, permission_id)` |
| org_nodes | `(company_id, id)` |
| members | `(company_id, email)` |
| models | `(company_id, call_type)` |
| tenant_background_state | `(company_id)` |

---

## BOOTSTRAP_MODE

| 值 | 行为 | 场景 |
|----|------|------|
| `prod` | `bootstrap.ApplyBootstrap` + `ReconcilePresetRoles` | selfhosted / SaaS 首次部署 |
| `minimal` | 同 prod，但使用精简 demo snapshot（少量部门/成员/key） | 本地开发快速启动 |
| `demo` | `bootstrap.ApplyBootstrap` + `ReconcilePresetRoles` + `demo.ApplyDemo` | 本地开发、完整演示 |
| `none` | 不做数据初始化，空 DB 报错 | 外部迁移工具管理 |

### prod vs minimal vs demo

- **prod** — 纯生产 bootstrap：只写必要的 currencies/companies/permissions/roles/org root/admin。数据最小化，适合真实部署。
- **minimal** — 在 prod 基础上叠加一份精简 demo snapshot（单部门、少量成员、一个 key），方便开发调试但不需完整假数据。
- **demo** — 在 prod 基础上叠加完整 demo 数据（多部门、多成员、预算、用量、审计日志等），用于产品演示。

---

## Preset Roles 策略

### 现状分析

preset roles 有两个写入路径：
1. **SaaS 动态创建** — `company.provisionCompany` → `defaultCompanyRoles(normalizer)` → `SetRoles`
2. **Seed 写入** — `seed/apply/seed_core.go` → `insertSeedRoles` → `INSERT INTO roles`

两者用同样的 manifest 数据源（`infra/permission/manifest.json` 里的 `presetRoles`），但入口不同。

### 问题

- **权限升级**：加了新 permission 后，已有 preset role 的 grants 不会自动扩展（`ON CONFLICT DO NOTHING`）
- **SaaS vs Selfhosted 分裂**：SaaS 每次建公司动态算权限（永远最新），selfhosted bootstrap 后 grants 就冻结了

### 最终方案：Reconcile-on-Start

**不在 bootstrap 里写 role_permission_grants。** 改为启动时执行一个轻量 reconcile：

```go
// seed/bootstrap/roles.go

// ReconcilePresetRoles ensures preset roles' permission grants match the current manifest.
// Runs on every startup (not just first boot). Idempotent, incremental, never deletes.
func ReconcilePresetRoles(ctx context.Context, exec TableWriter, companyID uuid.UUID) error {
    manifest := permission.MustManifest()
    for roleName, capabilities := range manifest.PresetRoles {
        roleID := getOrInsertPresetRole(ctx, exec, companyID, roleName)
        currentGrants := selectCurrentGrants(ctx, exec, roleID)
        expectedGrants := resolvePermissionIDs(capabilities, manifest)
        toAdd := difference(expectedGrants, currentGrants)
        for _, permID := range toAdd {
            insertGrant(ctx, exec, companyID, roleID, permID)  // ON CONFLICT DO NOTHING
        }
        // 注意：不删除 currentGrants 里多余的（可能是 admin 手动添加的）
    }
    return nil
}
```

**核心规则**：
- **只增不删** — reconcile 只添加 manifest 要求的 grants，不删除多余的（尊重手动自定义）
- **每次启动执行** — 不需要版本号追踪，manifest 本身就是最新状态的声明
- **所有部署模式统一** — SaaS 新建公司时仍走 `provisionCompany`，但已有公司的 preset roles 由 reconcile 保鲜
- **代价极低** — 只有 SELECT + 若干 INSERT（ON CONFLICT DO NOTHING），毫秒级
- **只管 preset roles** — custom roles（如"预算审批员"）不参与 reconcile，它们由用户手动创建和管理

### Custom Roles 不参与 Reconcile

manifest.json 的 `presetRoles` 只包含 preset 角色。像"预算审批员"这样的 `Type: "custom"` 角色不在 manifest 中声明，也不由 reconcile 管理。它们的 grants 完全由 admin 手动维护（或 demo seed 创建）。

### Reconcile 时机

```
store/postgres.New(ctx, cfg)
  ├─ applySchema(ctx, pool)
  └─ seed.Init(ctx, pool, store, cfg)
       ├─ bootstrap.ApplyBootstrap(ctx, pool, appCfg, bootstrapCfg)
       ├─ bootstrap.ReconcilePresetRoles(ctx, pool, allCompanyIDs)  // ← 每次启动
       └─ if demo && empty: demo.ApplyDemo(ctx, store, cfg)
```

### 对比其他方案

| 方案 | 优点 | 缺点 |
|------|------|------|
| ❌ DB migration 管理 grants | 精确控制 | 每加一个 permission 就要写 migration SQL，容易忘 |
| ❌ bootstrap 里 `ON CONFLICT DO UPDATE` | 简单 | 会覆盖 admin 自定义的 grants |
| ❌ 版本号追踪 | 只在升级时执行 | 需要维护 seed_version 表，复杂度高 |
| ✅ Reconcile-on-Start (只增) | 自愈、无状态、尊重自定义 | 多余 grants 不清理（但这是 feature not bug） |

---

## 依赖方向

```
store/postgres       ──→  seed/                          ✅ (调 seed.Init)
seed/                ──→  seed/bootstrap                  ✅
seed/                ──→  seed/demo                       ✅ (条件调用)
seed/demo            ──→  seed/bootstrap                  ✅ (读 DefaultConfig)
seed/bootstrap       ──→  internal/infra/permission       ✅ (读 manifest)
seed/bootstrap       ──→  internal/store (types)          ✅ (用 Snapshot 等 struct)
seed/bootstrap       ──→  internal/domain/types           ✅ (Permission, Role struct)
seed/bootstrap       ──→  internal/domain/grants          ✅ (Role name 常量)
internal/domain/*    ──✗  seed/*                         ❌ 禁止
internal/pkg/*       ──✗  seed/*                         ❌ 禁止
internal/infra/*     ──✗  seed/*                         ❌ 禁止
tests/*              ──→  seed/demo/contract              ✅
tests/*              ──→  seed/bootstrap                  ✅
```

### 依赖方向说明

bootstrap 允许 import internal 的好处：
- **不需要维护第三份 manifest.json 副本** — 直接用 `internal/infra/permission.MustManifest()`
- **不需要重复定义 types** — 直接用 `internal/domain/types.Permission` 等 struct
- **code-gen 输出只需两份**（contracts → internal/infra/permission/manifest.json + keys.go）

关键约束是**单向**的：internal 不能 import seed，避免循环依赖和初始化耦合。

---

## Permissions 数据源统一

**单一真相来源**：`packages/contracts/permission/manifest.json`

manifest 扩展 schema（加入 `permissionNames` 供 bootstrap 写 permissions 表的 name/group 列）：

```json
{
  "version": 1,
  "capabilities": [...],
  "permissionIdMap": {"p-1": "org:structure", ...},
  "permissionNames": {
    "p-1": {"name": "组织架构管理", "group": "组织"},
    "p-2": {"name": "成员管理", "group": "组织"},
    ...
  },
  "presetRoles": {
    "超级管理员": ["*"],
    "普通成员": ["self:keys", "self:approval"],
    ...
  }
}
```

code-gen 输出：
1. `internal/infra/permission/manifest.json` — 运行时 capability 解析 + bootstrap 共用
2. `internal/infra/permission/keys.go` — 运行时常量

bootstrap 通过 `permission.MustManifest()` 获取数据，包括 `permissionNames`，用于写 permissions 表。

不再需要第三份 `seed/bootstrap/manifest.json`。

---

## 调用链

```
store/postgres.New(ctx, cfg)
  ├─ applySchema(ctx, pool)                          // DDL only
  ├─ <构造 Store struct>
  └─ seed.Init(ctx, pool, store, cfg)                // 数据初始化
       │
       ├─ bootstrapCfg := bootstrap.LoadConfig(cfg.BootstrapConfigPath)
       ├─ bootstrap.ApplyBootstrap(ctx, pool, cfg, bootstrapCfg)
       │     // currencies, companies, permissions, roles, org, admin, models
       │     // 全部 ON CONFLICT DO NOTHING，幂等
       │
       ├─ bootstrap.ReconcilePresetRoles(ctx, pool, companyIDs)
       │     // 每次启动：确保 preset roles grants 包含 manifest 最新权限
       │     // 直接调用 permission.MustManifest() 获取 preset role 定义
       │
       └─ if cfg.BootstrapIsDemo() && isDatabaseEmpty:
              demo.ApplyDemo(ctx, store, cfg)
                   // demo 数据用 store.Store（需要 repo 方法）
```

### 时序说明

- **bootstrap** 只用 `TableWriter`（`pool.Exec`），不需要完整 Store
- **reconcile** 同上，raw SQL + `permission.MustManifest()`
- **demo** 用完整 `store.Store`，因为需要 repo 方法（usage、ledger 等复杂写入）
- Store struct 构造在 bootstrap 之后、demo 之前

---

## 常量归属

### `seed/bootstrap/` — config 默认值

不再单独暴露常量。默认值内聚在 `DefaultConfig()` 里：

```go
func DefaultConfig() Config {
    return Config{
        Version: 1,
        Company: CompanyConfig{Name: "My Company"},
        Billing: BillingConfig{
            Currency:     "CNY",
            QuotaPerUnit: 500_000,
        },
    }
}
```

测试/demo 需要引用 quota_per_unit 时：`bootstrap.DefaultConfig().Billing.QuotaPerUnit`。

### `internal/pkg/common/` — 运行时常量（保留）

```go
const DefaultBillingCurrency = "CNY"       // company 无 billing_currency 时 fallback
const DefaultPersonalBudget = 0            // 新成员默认个人预算
const NewAPIGroupPrefix = "dept-"          // NewAPI group 命名
const ModelNotInDeptMessage = "..."        // 网关拒绝消息
// QuotaFromAmount, QuotaToDisplay — 纯计算函数
```

### 删除

| 删除项 | 原因 |
|--------|------|
| `common.DefaultQuotaPerUnit` | 运行时从 currencies 表读；种子值在 bootstrap config |
| `billing.DefaultQuotaPerUnit()` | 运行时走 `resolveChargeRate` |

---

## `DefaultQuotaPerUnit` 消除

| 引用位置 | 改为 |
|---------|------|
| `seed/` 各处 | `bootstrap.DefaultConfig().Billing.QuotaPerUnit` |
| `store/postgres/bootstrap.go` | 删除（逻辑移入 seed/bootstrap） |
| `billing/lot.go` DefaultQuotaPerUnit() | 删除 |
| `billing/trial.go` 硬编码 | 调用方传入 `quotaPerUnit` 参数（见下文） |
| 测试文件 | `bootstrap.DefaultConfig().Billing.QuotaPerUnit` |

### trial.go 改造

`SeedTrialCredit` 不再自己硬编码 `common.DefaultQuotaPerUnit`。改为增加 `quotaPerUnit` 参数：

```go
// Before:
func SeedTrialCredit(ctx context.Context, st CreditStore, companyID uuid.UUID, trialQuota int64) error {
    ppu := common.DefaultQuotaPerUnit  // ← 硬编码
    ...
}

// After:
func SeedTrialCredit(ctx context.Context, st CreditStore, companyID uuid.UUID, trialQuota int64, quotaPerUnit int64) error {
    // quotaPerUnit 由调用方从 currencies 表查出后传入
    ...
}
```

调用方（注册流程）先调 `billing.ResolveCompanyChargeRate(ctx, store, companyID)` 拿到 ppu，然后传入。这样 trial.go 不需要自己查 DB，也不硬编码。

---

## demo/ 数据（开发环境叠加）

在 bootstrap 之上追加：

| 数据 | 说明 |
|------|------|
| 多级部门（8 个） | 完整组织树 |
| 成员（11+） | 多角色、多部门 |
| 预算树 | 部门/项目预算配置 + 消费记录 |
| Platform Keys（6 个） | member/project 各类型 |
| Provider Keys | |
| Models catalog | 10 个模型 |
| Model allowlist | 部门-模型白名单 |
| 审批记录 | |
| 充值 + Lots | 模拟钱包 |
| Usage | usage buckets + ledger |
| 审计日志 | |

---

## 验证标准

- [ ] `BOOTSTRAP_MODE=prod` + 空 DB + 默认 config → 启动成功，admin 可登录
- [ ] `BOOTSTRAP_MODE=prod` + 自定义 bootstrap.yaml → 公司名/admin/货币/模型按配置写入
- [ ] `BOOTSTRAP_MODE=minimal` + 空 DB → 精简 demo 数据可用
- [ ] `BOOTSTRAP_MODE=demo` + 空 DB → 完整 demo 数据可展示
- [ ] `BOOTSTRAP_MODE=none` + 空 DB → 报错
- [ ] 重复启动 → bootstrap 幂等无副作用；reconcile 补齐新 permission grants
- [ ] 手动给 preset role 额外加的 grants 不被 reconcile 删除
- [ ] custom roles（如"预算审批员"）不受 reconcile 影响
- [ ] `common` 包不再暴露 `DefaultQuotaPerUnit`
- [ ] `internal/domain/*`、`internal/pkg/*`、`internal/infra/*` 不 import `seed/`
- [ ] `seed/bootstrap/` 可 import `internal/`（单向依赖）
- [ ] `billing/trial.go` 不硬编码 QuotaPerUnit（改为参数注入）
- [ ] manifest `permissionNames` 字段由 contracts code-gen 产出
- [ ] 所有测试通过

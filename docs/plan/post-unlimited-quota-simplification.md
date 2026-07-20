# 架构分析：Hardcode UnlimitedQuota 后的系统简化方案

> 日期：2026-07-20  
> 前提：适配层已 hardcode `UnlimitedQuota: true`，domain 层不再感知 NewAPI token quota 概念  
> 关联：docs/adr/newapi-wallet-quota-sync.md

---

## 一、当前状态快照

### 已完成的变更

| 变更 | 说明 |
|------|------|
| `adminport.CreateTokenInput` | 删除了 `RemainQuota`、`UnlimitedQuota` 字段 |
| `adminport.UpdateTokenInput` | 删除了 `RemainQuota`、`UnlimitedQuota` 字段 |
| `admin_port_adapter.go` | hardcode `UnlimitedQuota: true`（适配层） |
| `platformkey/create.go` | 不再传 `RemainQuota`/`UnlimitedQuota` |
| `platformkey/update.go` | 只同步 status/model_limits/group，不传 quota |

### 还留着但需要处理的组件

| 组件 | 当前行为 | 状态 |
|------|---------|------|
| **Rebalance job + worker** | 遍历 mappings → `RefreshPlatformKeyCombined`（更新 PG combined_key_remain） | ✅ 保留 — Gateway precheck 依赖 |
| **combined_key_summaries 表** | 存每个 key 的 remain 值，gateway precheck 读取 | ✅ 保留 — 预算限额数据源 |
| **SyncUpdatePlatformKey 中的 budget 计算** | `LoadBudgetContext` + `ComputeRemainForMapping` + `UpdateBatch` | ⚠️ 冗余 — 调用方已自行调 `RefreshPlatformKeyCombined` |
| **SyncCreatePlatformKey 中的 budget 计算** | `LoadBudgetContext` + `ComputeRemainForMapping`（赋值给已删除的 RemainQuota） | ⚠️ 死代码 — 无任何 side effect |
| **bootstrap.go ensureWalletUserQuota** | 用 MaxInt32 TopUp user quota | ⚠️ hacky — 需要改为 TopUp to wallet_remain |
| **NewAPI user quota** | 新 user = MaxInt32 / 旧 user = 0 | ⚠️ 需要正确的同步策略 |

---

## 二、核心问题：NewAPI User Quota 怎么办

Token 是 unlimited 了，但 **NewAPI 仍然检查 user-level quota**。必须解决。

### 方案对比

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| A: 充值时 TopUp | CreditFromLot 后 TopUp delta | 精确、语义清晰 | billing service 需要 adminport 依赖 |
| B: bootstrap 一次性补齐 | 启动时 TopUp 到 wallet_remain | 简单 | 两次启动之间的充值不会反映到 NewAPI |
| C: A + B 组合 | 充值 TopUp + 启动补齐 | 覆盖全场景 | 稍微多一点代码 |

**结论：采用 C（ADR 已描述）**

---

## 三、可以删除/简化的组件

### 3.1 ✅ 已确认可以删除

| 组件 | 原因 |
|------|------|
| `SyncUpdatePlatformKey` 中的 budget 计算 | update.go 不再需要 `ComputeRemainForMapping` — 因为它不再同步 remain 到 NewAPI |
| `update.go` 对 `CombinedKeySummaries().UpdateBatch` 的调用 | 之前是"顺便"更新 local remain，但 Rebalance job 独立负责这个 |
| `create.go` 中的 budget 加载路径 | 创建 token 不需要计算 remain 了（之前用来设 RemainQuota） |
| `pkgbudget` / `models` / `rules` 加载（在 create/update 的 quota 计算分支） | 只有 model_limits 还需要它 |

### 3.2 ✅ 可以简化

| 组件 | 简化方向 |
|------|---------|
| `platformkey/update.go` | 删除 `pkgbudget.LoadBudgetContext`、`ComputeRemainForMapping`、`CombinedKeySummaries().UpdateBatch`。只保留：读 key → 解析 model limits → 调 UpdateToken(status/models/group) |
| `platformkey/create.go` | 删除 `pkgbudget.LoadBudgetContext`、`ComputeRemainForMapping`、`remain` 变量。只保留：解析 model limits → 调 CreateToken → persist mapping |
| `RebalanceService.rebalanceKey` | 这个**不能删** — 它负责更新 `combined_key_remain`，Gateway precheck 依赖它 |
| `bootstrap.go` | 用 ADR 方案替换 MaxInt32 hacky 逻辑 |

### 3.3 ❌ 不能删除（仍有用）

| 组件 | 为什么保留 |
|------|-----------|
| **Rebalance job** | 更新 `combined_key_summaries`，Gateway precheck 读这个值决定是否拒绝 |
| **combined_key_summaries 表** | Gateway 预算限额的数据源 |
| **combined_key_remain 递减（ingest 路径）** | `DecrementBatch` 在 ingest 事务内实时扣减 remain |
| **Redis budget cache** | Gateway 高频 precheck 的读缓存层 |
| **SyncUpdatePlatformKey 本身** | 仍然需要同步 status/model_limits/group 到 NewAPI |
| **SyncCreatePlatformKey** | 仍然需要在 NewAPI 创建 token |

---

## 四、`update.go` 清理后的目标形态

```go
func SyncUpdatePlatformKey(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID, targetActive *bool) error {
    // 1. 读 key + mapping
    mapping := ...
    key := ...
    
    // 2. 解析 model limits（仍需 departments/rules/models）
    effectiveCallTypes := resolveModelLimits(...)
    
    // 3. 调 NewAPI UpdateToken（只传 status/model_limits/group）
    d.Client.UpdateToken(ctx, req)
    
    // 4. 更新 mapping sync status
    d.Mappings.UpdateMappingSync(...)
    
    // 注意：不再计算/更新 combined_key_remain
    // combined_key_remain 由 Rebalance job 独立维护
}
```

### 问题：update.go 删掉 `CombinedKeySummaries().UpdateBatch` 后，Rebalance 覆盖得到吗？

**是的。** 每次 budget 变更（成员额度修改、部门额度修改、项目变更）都会 enqueue Rebalance job，Rebalance 会重新计算 combined_key_remain。唯一"实时性"差异是：

- 之前：UpdateToken 时**顺便**刷新 combined_key_remain
- 之后：等 Rebalance job 运行（秒级延迟）

这个延迟可接受——model_limits 变更不影响预算额度。

### 问题：create.go 删掉 budget 计算后，新 key 的 combined_key_remain 谁初始化？

目前 `create.go` 不写 combined_key_summaries（只是计算了 `remain` 赋给已删除的 `RemainQuota` 字段）。新 key 的 combined_key_remain 初始化由：
1. **首次 ingest** — absoluteRecompute fallback
2. **bootstrap** — reconcileSyncedPlatformKeyMappings
3. **budget 变更** — Rebalance job

所以删掉 create.go 的 budget 计算 **没有功能损失**。

---

## 五、`create.go` 清理后的目标形态

```go
func TrySyncCreate(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID) (string, error) {
    // 1. 读 key + mapping（已存在的提前返回）
    existing := ...
    key := ...
    
    // 2. 解析 model limits（需 departments/rules/models）
    effectiveCallTypes := resolveModelLimits(...)
    
    // 3. Ensure group
    d.Client.EnsureGroup(...)
    
    // 4. 调 NewAPI CreateToken（无 quota 字段）
    token := d.Client.CreateToken(ctx, CreateTokenInput{
        UserID, Name, ModelLimitsEnabled, ModelLimits, Group, ExpiredTime
    })
    
    // 5. Persist key secret + mapping
    persistPlatformKeySecret(...)
    d.Mappings.UpdateMappingSync(...)
}
```

**删除的加载项：**
- `pkgbudget.LoadBudgetContext` — 不再需要
- `pkgbudget.OpenDepartmentPeriod` — 不再需要
- `pkgbudget.ComputeRemainForMapping` — 不再需要
- `newapiunits.ToNewAPIUnits` — 已删除
- `_ = remain` 那一行

**保留的加载项：**
- `common.LoadDepartments` — model limits 需要
- `common.LoadRoutingRules` — model limits 需要
- `d.Store.Models().Models` — model limits 需要

---

## 六、NewAPI User Quota 同步策略（详见 ADR）

| 时机 | 操作 | 实现位置 |
|------|------|---------|
| 充值成功后 | `TopUp(walletUserID, quotaGranted)` | `billing/lot_confirm.go` 每个路径的 CreditFromLot 后 |
| 应用启动 | `TopUp(walletUserID, max(0, walletRemain - currentQuota))` | `provision/bootstrap.go` |
| 消费 | **不需要操作** — NewAPI 消费时自动扣 user quota | — |

---

## 七、副作用分析总结

| 变更 | 风险 | 结论 |
|------|------|------|
| create.go 删 budget 计算 | 新 key 没有初始 combined_key_remain | 安全 — absoluteRecompute/rebalance/key创建后的 `RefreshPlatformKeyCombined` 覆盖 |
| update.go 删 combined_key_remain 更新 | model_limits 变更后 remain 不立即刷新 | 安全 — 调用方（keys domain）已自行调 `RefreshPlatformKeyCombined`；update.go 的是冗余 |
| bootstrap TopUp to wallet_remain | 启动前的消费可能让 NewAPI quota < tokenjoy wallet_remain | 不可能 — NewAPI 消费和 tokenjoy ingest 扣同样的值 |
| 充值后 TopUp 失败 | NewAPI user quota < wallet_remain | 可接受 — 下次充值/启动补齐；Gateway precheck 仍正常 |
| 充值后 TopUp 重复执行 | NewAPI user quota > wallet_remain | 安全 — Gateway precheck 在 NewAPI 之前拦截；多余 quota 不造成安全风险 |

### 调用链验证（update.go 的 combined 更新是冗余的证据）

| 调用 `SyncUpdatePlatformKey` 的地方 | 调用后是否自行刷新 combined_key_remain |
|-------------------------------------|--------------------------------------|
| `keys/platform_key_actions.go`（toggle enabled） | ✅ 直接调 `RefreshPlatformKeyCombined` |
| `keys/platform_key_newapi.go`（update budget/models） | ✅ 当 budget 变更时调 `RefreshPlatformKeyCombined` |
| `modellimits/modellimits.go`（路由规则变更） | model_limits 变更不影响预算额度 |
| `provision/bootstrap.go`（启动 reconcile） | ✅ Rebalance job 在之后全量刷新 |

### 性能影响

| 变更 | 对性能的影响 |
|------|------------|
| create.go 删 budget 计算 | **微量提升** — 少做 5~6 次 PG 查询（仅创建 key 时，极低频） |
| update.go 删 budget 计算 + combined 更新 | **提升** — 少做 `LoadBudgetContext`(5 PG 查询) + `UpdateBatch`(1 PG 写)，每次 key toggle/model 变更时 |
| 充值后 TopUp HTTP 调用 | **+50~200ms**（仅在充值时，一天几次到几十次，非热路径） |

---

## 八、改动清单

### Phase 1：清理 create.go / update.go（纯删除）

| 文件 | 改动 |
|------|------|
| `platformkey/create.go` | 删除 `pkgbudget.LoadBudgetContext`、`OpenDepartmentPeriod`、`ComputeRemainForMapping`、`_ = remain` |
| `platformkey/update.go` | 删除 `pkgbudget.LoadBudgetContext`、`OpenDepartmentPeriod`、`ComputeRemainForMapping`、`CombinedKeySummaries().UpdateBatch` |
| `platformkey/update.go` | 删除 `pkgbudget` import |

### Phase 2：修复 bootstrap.go（替换 MaxInt32）

| 文件 | 改动 |
|------|------|
| `provision/bootstrap.go` | 新建时 Quota = `wallet_remain`（当前公司余额）；已有 user 时 TopUp 差额 |

### Phase 3：充值后 TopUp（新增功能）

| 文件 | 改动 |
|------|------|
| `billing/service.go` | `service` struct 加 `adminClient adminport.Port` |
| `billing/wallet_topup.go`（新） | `topUpNewAPIQuota(ctx, companyID, delta)` |
| `billing/lot_confirm.go` | 4 个充值路径 CreditFromLot 后调 topUp |

### Phase 4：测试

| 文件 | 改动 |
|------|------|
| 现有 bootstrap test | 断言调用 TopUp with wallet_remain delta |
| 新增 billing test | 验证充值后 TopUp 被调用 |

---

## 九、不改的东西（明确排除）

- Rebalance job / worker — 仍然需要（维护 `combined_key_remain`）
- combined_key_summaries 表 — 仍然需要（Gateway precheck 数据源）
- Gateway precheck 逻辑 — 不变（wallet_remain + combined_key_remain 两层检查）
- Redis budget cache — 不变（高频 precheck 读缓存）
- ingest 路径的 DecrementBatch — 不变（实时扣减 remain）
- overrun 检测逻辑 — 不变（remain ≤ 0 时触发 overrun job）
- `RefreshPlatformKeyCombined` — 不变（key 创建/修改/rebalance 时调用）
- `ComputeRemainForMapping` / `ComputeGatewaySummaryUpdates` — 不变（Rebalance 链路依赖）

---

## 十、NewAPI 两层限额的完整语义

```
请求进入
    │
    ▼
┌─────────────────────────────────┐
│ tokenjoy Gateway precheck (PG)  │  ← 业务限额控制面
│  • wallet_remain > 0            │
│  • combined_key_remain > 0      │
└────────────────┬────────────────┘
                 │ 放行
                 ▼
┌─────────────────────────────────┐
│ NewAPI                          │  ← 物理止损层
│  • user quota > 0  ← 由 TopUp 保证 ≥ wallet_remain
│  • token unlimited ← hardcode true
└─────────────────────────────────┘
```

两层防线的不变量：`NewAPI user quota ≥ tokenjoy wallet_remain`

- 正常运行时始终成立（两者扣减量相等，TopUp 增量相等）
- 唯一不一致窗口：TopUp HTTP 失败时（概率极低，下次充值/重启自动修复）

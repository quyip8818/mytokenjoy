# Currency 管理

平台通过 `currencies` 表定义支持的计费币种及 quota 换算率（`quota_per_unit`）。本文描述 currency 的完整生命周期：数据模型、管理入口、SaaS→Local 同步、对计费的影响。

**相关：** [Backend-计费模式.md](./Backend-计费模式.md)（计费全流程） · [plan/catalog-currency-sync.md](./plan/catalog-currency-sync.md)（同步方案设计）

---

## 1. 数据模型

```sql
CREATE TABLE currencies (
    currency         CHAR(3) PRIMARY KEY,      -- ISO 4217, e.g. CNY/USD
    quota_per_unit   BIGINT NOT NULL,           -- 1 单位货币 = ? quota
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

FK 约束：`companies.billing_currency REFERENCES currencies(currency)` — 币种不可物理删除。

---

## 2. 核心概念

| 概念 | 说明 |
|------|------|
| `quota_per_unit` (QPU) | 1 单位法币对应多少内部 quota。默认 `1 CNY = 500,000 quota` |
| 快照 | 充值时 QPU 锁入 lot，后续 QPU 变更不影响已有 lot |
| enabled | 禁用后该币种不可用于新充值，已有 lot 不受影响 |
| SOT | SaaS 模式下 currencies 表由 SaaS 实例维护，Local 通过 CatalogSync 拉取 |

---

## 3. 运行模式

### 3.1 SaaS（`SupportSaas=true`）

SaaS 是 currencies 的 Source of Truth。平台管理员通过 UI 管理，每次写操作自动 bump `catalog.currencies_version`。

```
平台管理员 → CRUD API → currencies 表 → bump version
                                         ↓
                               CatalogSync (5min 周期)
                                         ↓
                               Local 实例拉取并覆盖
```

### 3.2 Local（`CatalogSyncEnabled=true`）

Local 不直接编辑 currencies，由 CatalogSync 从 SaaS 同步。首次启动时 seed 写入 CNY 作为 fallback（sync 执行前可用），sync 成功后 upsert 覆盖。

### 3.3 独立部署（`SupportSaas=false, CatalogSyncEnabled=false`）

currencies 由 seed 初始化，后续手动修改数据库或扩展管理接口。

---

## 4. 管理 API（SaaS 平台管理员）

所有写操作需 `PLATFORM_MANAGE` 权限，每次成功后自动 `Increment(catalog.currencies_version)`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/platform/currencies` | 列出所有币种（含 disabled） |
| POST | `/api/platform/currencies` | 新增币种（code + quotaPerUnit） |
| PUT | `/api/platform/currencies/{code}` | 修改 QPU |
| PATCH | `/api/platform/currencies/{code}/status` | 启用/禁用 |

### 约束

- 币种代码：3 位大写字母（`^[A-Z]{3}$`）
- QPU：正整数
- 禁用保护：被非归档企业引用的币种不可禁用（返回 409）
- 不可删除：FK 约束 + 已有 lot 引用

---

## 5. CatalogSync 同步

### 5.1 SaaS 暴露的 Sync Endpoint

| 端点 | 认证 | 说明 |
|------|------|------|
| `GET /api/platform/sync/versions` | 无 | 返回 `{ models, pricing, currencies }` 版本号 |
| `GET /api/platform/sync/catalog/currencies` | 无 | 返回 `{ version, data: [{code, quotaPerUnit}] }`，仅含 enabled 币种 |

### 5.2 Local 同步流程

```
Executor.Execute() — 每 5 分钟
  ├─ FetchVersions → remote.Currencies
  ├─ 比较 local system_settings["catalog.currencies_version"]
  │   相同 → 跳过
  │   不同 → FetchCurrencies
  ├─ WithTx:
  │   ├─ Upsert 所有远端币种 (enabled=true)
  │   └─ Disable 不在列表中的 (enabled=false)
  └─ Set local version = resp.Version
```

### 5.3 安全性

- 不做物理 DELETE（FK 约束）
- 被引用的币种 disable 后：已有 lot 不受影响，新充值被 `lookupQuotaPerUnit` 拒绝
- 全量覆盖 O(10) 条，无性能问题

---

## 6. 对计费的影响

| 场景 | 行为 |
|------|------|
| QPU 变更 | 仅影响后续充值。已有 lot 的 QPU 是快照，不变 |
| 币种禁用 | `lookupQuotaPerUnit` 返回错误，新充值/overdraft 被拒 |
| 新增币种 | 需先在 SaaS 创建，企业才能将 `billing_currency` 设为该币种 |
| 改企业币种 | 旧 lot 保持原币种和 QPU，新充值用新币种 |

### QPU 快照时序

```
t0: currencies.CNY.QPU = 500,000
t1: 企业 A 充值 ¥100 → lot(QPU=500,000, quota=50,000,000)
t2: admin 改 QPU = 600,000
t3: 企业 A 再充 ¥100 → lot(QPU=600,000, quota=60,000,000)
t4: 消耗 lot1 时 display = quota / 500,000
    消耗 lot2 时 display = quota / 600,000
```

---

## 7. 前端 UI

路由：`/platform/currencies`（平台管理分组，仅 `PLATFORM_MANAGE` 可见）

功能：
- 表格展示所有币种（代码 / QPU / 状态）
- 新增币种 Dialog（3 位大写代码 + QPU）
- 编辑 QPU Dialog（带影响提示）
- 启停切换（禁用时二次确认）

---

## 8. 代码地图

| 层 | 文件 | 职责 |
|---|------|------|
| Schema | `store/postgres/schema.sql` | currencies 表定义 |
| Seed | `seed/apply/seed_core.go` | CNY 初始化 + catalog version |
| Store 接口 | `store/billing_repo.go` | Currency struct + 6 个方法 |
| Store 实现 | `store/postgres/billing_repo_currency.go` | PG 实现 |
| 计费解析 | `domain/billing/currency.go` | lookupQuotaPerUnit / ResolveCompanyChargeRate |
| Platform Handler | `http/handler/platform/currencies.go` | Sync endpoint + CRUD |
| CatalogSync Types | `integration/catalogsync/types.go` | CatalogCurrency / CatalogVersions |
| CatalogSync Client | `integration/catalogsync/client.go` | FetchCurrencies |
| CatalogSync Worker | `worker/catalogsync/execute.go` | syncCurrencies |
| Frontend API | `api/platform.ts` | PlatformCurrency + CRUD 方法 |
| Frontend Feature | `features/platform/currencies/` | hook + page shell |
| Frontend Route | `config/routes.ts` + `routes/platform/currencies.tsx` | 路由注册 + 页面入口 |

---

## 9. Seed 与初始状态

| 环境 | seed 行为 | catalog.currencies_version |
|------|-----------|---------------------------|
| SaaS | 插入 CNY + 写 version=1 | 1 |
| Local | 插入 CNY（fallback） | 0（不写，触发首次同步） |
| 独立 | 插入 CNY | 不涉及 |

---

## 10. 测试覆盖

| 测试文件 | 验证内容 |
|----------|---------|
| `tests/worker/catalogsync/currencies_sync_test.go` | 同步触发 / disable stale / 版本相同跳过 |
| `tests/handler/platform/currencies_test.go` | CRUD 生命周期 / FK 禁用保护 / sync endpoint 格式 |
| `tests/domain/billing/service_test.go` | 充值时 QPU 从 currencies 表读取 |
| `tests/integration/budget/multirate_lot_test.go` | QPU 变更后新 lot 用新值 |

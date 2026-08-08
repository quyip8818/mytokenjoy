# Currency 管理

平台通过 `currencies` 表定义支持的计费币种及 quota 换算率（`quota_per_unit`）。本文描述 currency 的完整生命周期：数据模型、管理入口、SaaS→Local 同步、对计费的影响。

**相关：** [Backend-计费模式.md](./Backend-计费模式.md)（计费全流程）

---

## 1. 数据模型

```sql
CREATE TABLE currencies (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    currency              CHAR(3) NOT NULL,            -- ISO 4217, e.g. CNY/USD
    quota_per_unit        BIGINT NOT NULL CHECK (quota_per_unit > 0),
    enabled               BOOLEAN NOT NULL DEFAULT TRUE,
    updated_by_user_id    UUID,                        -- 操作人 user.id
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_currencies_code_time ON currencies(currency, updated_at DESC);
```

**Append-only 设计：** 同一 `currency` code 可有多行。最新行（`updated_at` 最大）为当前生效状态，历史行保留作审计。所有写操作均为 INSERT，不做 UPDATE。

`companies.billing_currency` 为 `CHAR(3) NOT NULL DEFAULT 'CNY'`，无 FK 约束（应用层通过 `GetCurrency` 校验一致性）。

---

## 2. 核心概念

| 概念 | 说明 |
|------|------|
| `quota_per_unit` (QPU) | 1 单位法币对应多少内部 quota。bootstrap 默认 `1 CNY = 500,000 quota` |
| 快照 | 充值时 QPU 锁入 lot，后续 QPU 变更不影响已有 lot |
| enabled | 禁用后该币种不可用于新充值，已有 lot 不受影响 |
| SOT | SaaS 模式下 currencies 表由 SaaS 实例维护，Local 通过 CatalogSync 拉取 |
| append-only | 每次修改 INSERT 新行，"当前生效" = 同 code 下 `updated_at` 最大的行 |

---

## 3. 运行模式

### 3.1 SaaS（`SupportSaas=true`）

SaaS 是 currencies 的 Source of Truth。平台管理员通过 UI 管理，每次写操作 INSERT 新行并 bump `sync_versions` 表中 `(GlobalSyncVersion, "currencies")`。

```
平台管理员 → CRUD API (INSERT) → currencies 表新行 → bump version
                                                      ↓
                                            CatalogSync (10min 周期)
                                                      ↓
                                            Local 实例拉取并 INSERT (ON CONFLICT id DO NOTHING)
```

### 3.2 Local（非 SaaS 模式）

Local 不直接编辑 currencies，由 CatalogSync 从 SaaS 同步。首次启动时 seed 写入 CNY 作为 fallback（sync 执行前可用），sync 成功后追加新行。

### 3.3 独立部署（无 SAAS_PLATFORM_URL）

currencies 由 seed 初始化，后续手动修改数据库或扩展管理接口。CatalogSync 因无可用 URL 自动跳过。

---

## 4. 管理 API（SaaS 平台管理员）

所有写操作需 `PLATFORM_MANAGE` 权限，每次成功后自动 bump `sync_versions`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/platform/currencies` | 列出所有币种最新状态（含 disabled） |
| POST | `/api/platform/currencies` | 新增币种（code + quotaPerUnit） |
| PUT | `/api/platform/currencies/{code}` | 修改 QPU（INSERT 新行） |
| PATCH | `/api/platform/currencies/{code}/status` | 启用/禁用（INSERT 新行） |
| GET | `/api/platform/currencies/{code}/history` | 查看某币种变更历史 |

### 响应格式

```json
{
  "id": "uuid",
  "code": "CNY",
  "quotaPerUnit": 10000,
  "enabled": true,
  "updatedAt": "2025-07-01T10:00:00Z",
  "updatedByName": "张三"
}
```

### 约束

- 币种代码：3 位大写字母（`^[A-Z]{3}$`）
- QPU：正整数
- 禁用保护：被非归档企业引用的币种不可禁用（返回 409）
- 创建保护：已存在同 code 的行时返回 409

---

## 5. CatalogSync 同步

### 5.1 SaaS 暴露的 Sync Endpoint

| 端点 | 认证 | 说明 |
|------|------|------|
| `GET /api/platform/sync/versions` | sync token | 返回 `{ models, pricing, currencies, discounts, walletLots }` 版本号 |
| `GET /api/platform/sync/catalog/currencies` | 无 | 返回 `{ version, data: [{id, code, quotaPerUnit, enabled, updatedAt}] }`（最新行 per code） |

### 5.2 Local 同步流程

```
Executor.Execute() — 每 10 分钟
  ├─ FetchVersions → remote.Currencies
  ├─ 比较本地 sync_versions (GlobalSyncVersion, "currencies")
  │   相同 → 跳过
  │   不同 → FetchCurrencies
  ├─ 遍历 resp.Data:
  │   └─ InsertCurrencyFromSync(row) — ON CONFLICT (id) DO NOTHING
  └─ Set local version = resp.Version
```

### 5.3 幂等性

- UUID id 由 SaaS 端生成，Local 用 `ON CONFLICT (id) DO NOTHING` 实现幂等
- 重复 sync 不会产生冗余行
- 与 lots/orders 的同步模式统一

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
t0: currencies.CNY 最新行 QPU = 500,000
t1: 企业 A 充值 ¥100 → lot(QPU=500,000, quota=50,000,000)
t2: admin 改 QPU = 600,000 (INSERT 新行)
t3: 企业 A 再充 ¥100 → lot(QPU=600,000, quota=60,000,000)
t4: 消耗 lot1 时 display = quota / 500,000
    消耗 lot2 时 display = quota / 600,000
```

---

## 7. 前端 UI

路由：`/platform/currencies`（平台管理分组，仅 `PLATFORM_MANAGE` 可见）

功能：
- 表格展示所有币种当前状态（代码 / QPU / 状态 / 修改人 / 修改时间）
- 新增币种 Dialog（3 位大写代码 + QPU）
- 编辑 QPU Dialog（带影响提示）
- 启停切换（禁用时二次确认）
- **「历史」按钮** → Dialog（currency picker + 变更历史表格）

---

## 8. 代码地图

| 层 | 文件 | 职责 |
|---|------|------|
| Schema | `store/postgres/schema.sql` | currencies 表定义（UUID PK, append-only） |
| Seed | `seed/apply/seed_core.go` | CNY 初始化 |
| Store 接口 | `store/billing_repo.go` | Currency struct + InsertCurrency / GetCurrency / ListCurrencyHistory 等 |
| Store 实现 | `store/postgres/billing_repo_currency.go` | PG 实现（DISTINCT ON + INSERT） |
| 计费解析 | `domain/billing/currency.go` | lookupQuotaPerUnit / ResolveCompanyChargeRate |
| Platform Handler | `http/handler/platform/currencies.go` | Sync endpoint + CRUD + History |
| CatalogSync Types | `integration/catalogsync/types.go` | CatalogCurrency（含 id） |
| CatalogSync Client | `integration/catalogsync/client.go` | FetchCurrencies |
| CatalogSync Worker | `worker/catalogsync/execute.go` | syncCurrencies（InsertCurrencyFromSync） |
| Frontend API | `api/platform.ts` | PlatformCurrency + CRUD + listCurrencyHistory |
| Frontend Feature | `features/platform/currencies/` | hook + page shell + history dialog |
| Frontend Route | `config/routes.ts` + `routes/platform/currencies.tsx` | 路由注册 + 页面入口 |

---

## 9. Seed 与初始状态

Seed 直接 INSERT（不带 ON CONFLICT，因 append-only 无 unique 约束）：

| 路径 | 触发时机 | CNY QPU 默认值 |
|------|----------|---------------|
| `seed/bootstrap`（`insertCurrencies`） | 每次启动执行 | `500,000`（可经 YAML 覆盖） |
| `seed/apply`（`insertSeedCurrencies`） | SaaS 模式 demo 数据 | `10,000` |

---

## 10. 测试覆盖

| 测试文件 | 验证内容 |
|----------|---------|
| `tests/worker/catalogsync/currencies_sync_test.go` | 同步触发 / id 幂等 / 版本相同跳过 |
| `tests/worker/catalogsync/boot_sync_test.go` | 全通道 boot sync |
| `tests/domain/billing/service_test.go` | 充值时 QPU 从 currencies 表读取 |
| `tests/integration/budget/multirate_lot_test.go` | QPU 变更后新 lot 用新值 |

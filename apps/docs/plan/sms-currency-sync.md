# SMS → TokenJoy 汇率同步实施方案

> 从 sms-model-sync-v2.md 拆出的 currencies 分区。

**前置依赖**：`sms-model-sync-v2.md` 完成（Catalog.Versions 结构、River PeriodicJob、version 持久化基础设施均在 v2 中实现）。

---

## 1. 目标

SMS 作为汇率 SOT，TokenJoy 通过 catalog API 拉取 currencies 分区，全量替换本地 `currencies` 表。

## 2. 数据流

```
SMS currencies 表（运维配置）
        │
        ▼
GET /api/sync/catalog → response.currencies + response.versions.currencies
        │
        ▼
TokenJoy smssync Execute()
  → version 比对（system_settings）
  → 不同 → ReplaceCurrencies（upsert + disable stale）
  → 更新 system_settings 中的 currencies_version
```

## 3. SMS 侧改动

### 3.1 新建 currencies 表

```sql
-- sms/backend/schema.sql
CREATE TABLE IF NOT EXISTS currencies (
    currency       CHAR(3) PRIMARY KEY,
    quota_per_unit BIGINT NOT NULL CHECK (quota_per_unit > 0),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER set_updated_at BEFORE UPDATE ON currencies
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
```

### 3.2 domain/sync 类型扩展（已在 v2 中完成）

v2 已经定义了 `CatalogVersions` 和 `Catalog` 结构。本步骤只需加 `CatalogCurrency` 并在 `Catalog` 中包含 `Currencies` 字段。

```go
type CatalogCurrency struct {
    Code         string `json:"code"`
    QuotaPerUnit int64  `json:"quotaPerUnit"`
}
```

### 3.3 Store 接口扩展

```go
type Store interface {
    // ... v2 中已有的
    ListCurrenciesForSync(ctx context.Context) ([]CatalogCurrency, error)
    // GetPartitionVersion 已在 v2 中实现，需加 "currencies" case
}
```

`GetPartitionVersion("currencies")` 实现：`SELECT MAX(updated_at) FROM currencies`。

### 3.4 Service 组装

`GetCatalog()` 返回的 Catalog 补充 `Currencies` 字段和 `Versions.Currencies`。

---

## 4. TokenJoy 侧改动

### 4.1 integration/sms 类型扩展（已在 v2 中完成）

v2 已经定义了 `CatalogVersions` 和 `Catalog` 结构。本步骤只需加 `CatalogCurrency`。

```go
type CatalogCurrency struct {
    Code         string `json:"code"`
    QuotaPerUnit int64  `json:"quotaPerUnit"`
}
```

### 4.2 store 层：BillingRepository 加 ReplaceCurrencies

```go
// store/billing_repo.go
ReplaceCurrencies(ctx context.Context, currencies []Currency) error
```

实现策略（FK 安全）：

```sql
-- 1. Upsert 所有 SMS 中的币种
INSERT INTO currencies (currency, quota_per_unit, enabled, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (currency) DO UPDATE SET
    quota_per_unit = EXCLUDED.quota_per_unit,
    enabled = TRUE,
    updated_at = NOW();

-- 2. 不在列表中的标记 disabled（不 DELETE，因为 companies.billing_currency FK）
UPDATE currencies SET enabled = FALSE, updated_at = NOW()
WHERE currency != ALL($1) AND enabled = TRUE;
```

### 4.3 SyncTarget 接口加方法

```go
type SyncTarget interface {
    // ... existing
    ReplaceCurrencies(ctx context.Context, currencies []sms.CatalogCurrency) error
}
```

### 4.4 smssync worker Execute() 加 currencies 步骤

```go
// 在 Execute() 中加：
if catalog.Versions.Currencies != lastCurrenciesVersion {
    if err := w.target.ReplaceCurrencies(ctx, catalog.Currencies); err != nil {
        slog.Error("smssync: replace currencies failed", "error", err)
        // 不更新 version，下次重试
    } else {
        // 更新 system_settings 中的 sms_sync.currencies_version
        currenciesVersion = catalog.Versions.Currencies
    }
}
```

---

## 5. FK 约束注意

`companies.billing_currency REFERENCES currencies(currency)` — 不能 DELETE 被引用的币种。

方案：用 `enabled = FALSE` 代替 DELETE。`lookupQuotaPerUnit()` 已有 enabled 检查，disabled 的币种无法用于新充值。

---

## 6. 对计费的影响

- `quota_per_unit` 是 lot 级快照：充值时从 currencies 表读取存入 lot
- 汇率变更只影响**后续充值**，已有 lot 的 QPU 不变
- disabled 的币种：新充值会被 `lookupQuotaPerUnit` 拒绝

---

## 7. 实施步骤

| # | 改动位置 | 内容 |
|---|----------|------|
| 1 | `sms/backend/schema.sql` | 加 currencies 表 |
| 2 | `sms/backend/internal/domain/sync/service.go` | 加 CatalogCurrency，Catalog 补 Currencies 字段 |
| 3 | `sms/backend/internal/store/postgres/sync_store.go` | 实现 ListCurrenciesForSync + GetPartitionVersion("currencies") |
| 4 | `apps/backend/internal/integration/sms/client.go` | Catalog 加 CatalogCurrency + Currencies 字段 |
| 5 | `apps/backend/internal/store/billing_repo.go` | BillingRepository 加 ReplaceCurrencies |
| 6 | `apps/backend/internal/store/postgres/billing_repo_currency.go` | 实现 ReplaceCurrencies（upsert + disable stale） |
| 7 | `apps/backend/internal/worker/smssync/target.go` | SyncTarget 加 ReplaceCurrencies |
| 8 | `apps/backend/internal/worker/smssync/worker.go` | Execute() 加 currencies version 比对 + 同步 |

# Sync Version 统一设计（sync_versions）

> 状态：已实现  
> 前置：`apps/docs/design-company-discount.md`

---

## 1. 问题

所有 catalog sync version 都存在 `system_settings`（全局 key-value），存在两个问题：

1. `discounts_version` / `wallet_lots_version` 数据本质是 per-company，放全局导致噪音 sync
2. `system_settings` 表职责混杂（既存 version 又存 setup 配置），sync version 应该有自己的归属

| Version Key | 数据隔离 | 当前存储 | 问题 |
|---|---|---|---|
| `catalog.discounts_version` | per-company | 全局 | 公司 A 改 discount → 所有 Local fetch（N-1 空操作） |
| `catalog.wallet_lots_version` | per-company | 全局 | 公司 A 充值/消费 → 所有 Local fetch |
| `catalog.models_version` | 全局 | 全局 | 语义正确，但散落在 system_settings |
| `catalog.pricing_version` | 全局 | 全局 | 同上 |
| `catalog.currencies_version` | 全局 | 全局 | 同上 |

此外，`/api/platform/sync/catalog/discounts` 路由**未注册**，Local client 调用会 404。

---

## 2. 方案

### 新表：`sync_versions`

所有 catalog sync version 统一存此表。`company_id = uuid.Nil`（全零 sentinel UUID）代表全局 version。

```sql
CREATE TABLE IF NOT EXISTS sync_versions (
    company_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    type       TEXT NOT NULL,
    version    INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (company_id, type)
);
```

> ponytail: PG 的 PRIMARY KEY 不允许 NULL，用 sentinel UUID 代替 NULL 表示全局。不建 FK（sentinel 无对应 companies 行）。公司删除时由 handler 负责清理残留行。

Type 取值：
- `models` — 模型目录变更（全局）
- `pricing` — 定价变更（全局）
- `currencies` — 货币变更（全局）
- `discounts` — 模型折扣变更（per-company）
- `wallet_lots` — 钱包/Lot 变更（per-company）

### 从 `system_settings` 移除

- `catalog.models_version`
- `catalog.pricing_version`
- `catalog.discounts_version`
- `catalog.currencies_version`
- `catalog.wallet_lots_version`

---

## 3. Store 接口

```go
// store/sync_versions.go
package store

import (
    "context"
    "github.com/google/uuid"
)

// GlobalSyncVersion is the sentinel company_id for global (non-company-scoped) versions.
var GlobalSyncVersion = uuid.Nil // 00000000-0000-0000-0000-000000000000

type SyncVersionRepository interface {
    // Increment atomically bumps version+1, returns new value. Upserts if absent.
    Increment(ctx context.Context, companyID uuid.UUID, typ string) (int, error)
    // Set writes an exact version value (upsert). Used by Local sync executor.
    Set(ctx context.Context, companyID uuid.UUID, typ string, version int) error
    // Get returns current version (0 if not exists).
    Get(ctx context.Context, companyID uuid.UUID, typ string) (int, error)
    // GetVersions returns global + per-company versions in one call.
    // Returns two maps: global types → version, company types → version.
    GetVersions(ctx context.Context, companyID uuid.UUID) (global map[string]int, company map[string]int, err error)
}
```

Postgres 实现：

```go
func (r *repo) Increment(ctx context.Context, companyID uuid.UUID, typ string) (int, error) {
    var v int
    err := r.db.QueryRow(ctx, `
        INSERT INTO sync_versions (company_id, type, version)
        VALUES ($1, $2, 1)
        ON CONFLICT (company_id, type)
        DO UPDATE SET version = sync_versions.version + 1
        RETURNING version
    `, companyID, typ).Scan(&v)
    return v, err
}

func (r *repo) Set(ctx context.Context, companyID uuid.UUID, typ string, version int) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO sync_versions (company_id, type, version)
        VALUES ($1, $2, $3)
        ON CONFLICT (company_id, type)
        DO UPDATE SET version = $3
    `, companyID, typ, version)
    return err
}

func (r *repo) Get(ctx context.Context, companyID uuid.UUID, typ string) (int, error) {
    var v int
    err := r.db.QueryRow(ctx,
        `SELECT version FROM sync_versions WHERE company_id = $1 AND type = $2`,
        companyID, typ,
    ).Scan(&v)
    if errors.Is(err, pgx.ErrNoRows) {
        return 0, nil
    }
    return v, err
}

func (r *repo) GetVersions(ctx context.Context, companyID uuid.UUID) (map[string]int, map[string]int, error) {
    rows, err := r.db.Query(ctx,
        `SELECT company_id, type, version FROM sync_versions
         WHERE company_id IN ($1, $2)`,
        store.GlobalSyncVersion, companyID,
    )
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    global := make(map[string]int)
    company := make(map[string]int)
    for rows.Next() {
        var cid uuid.UUID
        var typ string
        var v int
        if err := rows.Scan(&cid, &typ, &v); err != nil {
            return nil, nil, err
        }
        if cid == store.GlobalSyncVersion {
            global[typ] = v
        } else {
            company[typ] = v
        }
    }
    return global, company, rows.Err()
}
```

### Store 聚合

```go
// store/store.go — Store interface 新增：
SyncVersions() SyncVersionRepository
```

---

## 4. 变更点清单

### 4.1 写入侧（SaaS）—— bump version

| 触发位置 | 改为 |
|----------|------|
| `models.go` `bumpPricingVersion` helper | `h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "pricing")` |
| `models.go` `PublishCatalog` | `h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "models")` |
| `currencies.go` `CreateCurrency` / `UpdateCurrency` / `ToggleCurrencyStatus` | `h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "currencies")` |
| `pricing.go` `SetCompanyDiscount` | `h.p.SyncVersions.Increment(ctx, companyID, "discounts")` |
| `domain/billing/lot_confirm.go` `syncWalletBestEffort` | `s.store.SyncVersions().Increment(ctx, companyID, "wallet_lots")` |
| `domain/usage/ingest.go` post-commit | `s.store.SyncVersions().Increment(ctx, companyID, "wallet_lots")` |

> ponytail: `bumpPricingVersion` 被 `SetGlobalPricing`、`SetModelPricing`、`CreateModel`（有价格时）共 3 处调用，改 helper 一处即覆盖全部。`bumpModelsCatalogVersion` 被 `CreateModel`、`UpdateModel`、`PublishCatalog` 调用，同理。currencies 3 处已逐个替换。

### 4.2 读取侧（SaaS）—— CatalogVersions endpoint

路由从 public 移至 sync-token-required 组。Handler 重写：

```go
func (h *Handler) CatalogVersions(w http.ResponseWriter, r *http.Request) {
    companyID, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
    if !ok {
        httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
        return
    }

    global, company, err := h.p.SyncVersions.GetVersions(r.Context(), companyID)
    if err != nil {
        httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
        return
    }

    response.JSON(w, http.StatusOK, map[string]int{
        "models":     global["models"],
        "pricing":    global["pricing"],
        "currencies": global["currencies"],
        "discounts":  company["discounts"],
        "walletLots": company["wallet_lots"],
    })
}
```

### 4.3 读取侧（SaaS）—— 各 Catalog data endpoint 的 version 字段

| Endpoint | 改为 |
|----------|------|
| `CatalogModels`（public） | `h.p.SyncVersions.Get(ctx, store.GlobalSyncVersion, "models")` |
| `CatalogPricing` | `h.p.SyncVersions.Get(ctx, store.GlobalSyncVersion, "pricing")` |
| `CatalogCurrencies`（public） | `h.p.SyncVersions.Get(ctx, store.GlobalSyncVersion, "currencies")` |
| `CatalogWalletLots` | `h.p.SyncVersions.Get(ctx, companyID, "wallet_lots")` |
| `CatalogDiscounts`（新增） | `h.p.SyncVersions.Get(ctx, companyID, "discounts")` |

### 4.4 新增 CatalogDiscounts endpoint

新文件 `handler/platform/catalog_discounts.go`：

```go
func (h *Handler) CatalogDiscounts(w http.ResponseWriter, r *http.Request) {
    companyID, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
    if !ok {
        httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
        return
    }

    rows, err := h.p.ModelDiscount.CurrentDiscounts(r.Context(), companyID)
    if err != nil {
        httputil.WriteError(w, err)
        return
    }

    // 复用已有 discountDTO（pricing.go 中定义）
    data := make([]discountDTO, len(rows))
    for i, row := range rows {
        data[i] = discountDTO{ModelType: row.ModelType, Discount: row.Discount}
    }

    v, _ := h.p.SyncVersions.Get(r.Context(), companyID, "discounts")
    response.JSON(w, http.StatusOK, map[string]any{"version": v, "data": data})
}
```

### 4.5 路由注册 (`handler.go`)

```go
// Public (全局数据，无需 company context)
r.Get("/sync/catalog/models", h.CatalogModels)
r.Get("/sync/catalog/currencies", h.CatalogCurrencies)

// Sync-token required (需 company context)
r.Group(func(r chi.Router) {
    r.Use(httpmiddleware.RequireSyncToken(h.p.Companies))
    r.Get("/sync/versions", h.CatalogVersions)
    r.Get("/sync/catalog/pricing", h.CatalogPricing)
    r.Get("/sync/catalog/discounts", h.CatalogDiscounts)     // 新增
    r.Get("/sync/catalog/wallet_lots", h.CatalogWalletLots)
})
```

### 4.6 Client 侧（Local `catalogsync/client.go`）

`FetchVersions` 改为带 auth：

```go
// Before
c.doGet(ctx, "/api/platform/sync/versions", false, &v)

// After
c.doGet(ctx, "/api/platform/sync/versions", true, &v)
```

### 4.7 Local 侧 Executor（`catalogsync/execute.go`）

Executor 从 remote 拿到 version 后，与本地 `sync_versions` 表比较。sync 完成后用 `Set` 写入 remote 的精确值：

```go
// 读本地 version（Local 视角一切 version 都存为 GlobalSyncVersion，因为只服务一个公司）
localPricing, _ := e.store.SyncVersions().Get(ctx, store.GlobalSyncVersion, "pricing")

// sync 完成后写入
_ = e.store.SyncVersions().Set(ctx, store.GlobalSyncVersion, "pricing", resp.Version)
```

> ponytail: Local 只服务一个公司，所有本地 version（包括 discounts、wallet_lots）统一用 `GlobalSyncVersion` 作为 companyID。Local 不需要 per-company 区分——它就是那个公司。

### 4.8 Platform deps

`httpdeps.Platform` 增加：

```go
SyncVersions store.SyncVersionRepository
```

---

## 5. 完整文件清单

| # | 文件 | 操作 | 说明 |
|---|------|------|------|
| 1 | `store/postgres/schema.sql` | 改 | 新增 `sync_versions` 表 |
| 2 | `store/sync_versions.go` | 新增 | 接口 + `GlobalSyncVersion` 常量 |
| 3 | `store/postgres/sync_versions_repo.go` | 新增 | PG 实现（Increment/Set/Get/GetVersions） |
| 4 | `store/store.go` | 改 | Store interface 加 `SyncVersions()` |
| 5 | `store/postgres/store.go` | 改 | 返回 repo 实例 |
| 6 | `store/postgres/tx.go` | 改 | txStore 实现 `SyncVersions()` |
| 7 | `http/deps/public.go` | 改 | Platform struct 加 `SyncVersions` 字段 + 赋值 |
| 8 | `http/handler/platform/handler.go` | 改 | 路由调整（versions 移入 auth 组 + 注册 discounts） |
| 9 | `http/handler/platform/models.go` | 改 | CatalogVersions 重写 + CatalogModels version 改源 + PublishCatalog 改源 + bumpPricingVersion 改源 + 删除旧常量 |
| 10 | `http/handler/platform/pricing.go` | 改 | SetCompanyDiscount 加 bump |
| 11 | `http/handler/platform/currencies.go` | 改 | 3 处 version bump 改 SyncVersions |
| 12 | `http/handler/platform/catalog_discounts.go` | 新增 | CatalogDiscounts sync endpoint |
| 13 | `http/handler/platform/catalog_lots.go` | 改 | version 改 SyncVersions |
| 14 | `http/handler/platform/catalog_pricing.go` | 改 | pricingVersion helper 改 SyncVersions |
| 15 | `domain/billing/lot_confirm.go` | 改 | wallet_lots bump 改 SyncVersions |
| 16 | `domain/usage/ingest.go` | 改 | wallet_lots bump 改 SyncVersions |
| 17 | `integration/catalogsync/client.go` | 改 | FetchVersions 带 auth |
| 18 | `worker/catalogsync/execute.go` | 改 | 本地 version 读写全部改 SyncVersions，删除旧 key 常量 |

### 测试需更新

| # | 文件 | 原因 |
|---|------|------|
| T1 | `tests/domain/usage/ingest_channel_split_test.go` | version bump 断言改用 SyncVersions |
| T2 | `tests/integration/localsaas/local_saas_test.go` | version 读取方式变更 |
| T3 | `tests/worker/catalogsync/boot_sync_test.go` | version 读写改 SyncVersions |
| T4 | `tests/worker/catalogsync/wallet_lots_sync_test.go` | 同上 |
| T5 | `tests/worker/catalogsync/currencies_sync_test.go` | 同上 |
| T6 | `tests/handler/platform/catalog_lots_test.go` | version 断言 |
| T7 | `tests/handler/platform/currencies_test.go` | fetchCurrenciesVersion helper 改源 |

---

## 6. 依赖注入路径

- `ingest.go` / `lot_confirm.go`：通过 `s.store.SyncVersions()` 访问（Store interface 已有）
- `handler/platform/*`：通过 `h.p.SyncVersions` 访问（Platform deps 注入）
- `worker/catalogsync/execute.go`：通过 `e.store.SyncVersions()` 访问

---

## 7. system_settings 表处理

迁移已完成。`system_settings` 不再存储任何 catalog version key。

**已确认清理项**：
- ✅ 旧常量 `catalogModelsVersionKey` / `catalogPricingVersionKey` / `catalogDiscountsVersionKey` / `catalogCurrenciesVersionKey` / `catalogWalletLotsVersionKey` — 已删除
- ✅ `seed_core.go` `seedCatalogVersions` — 已改写为写入 `sync_versions` 表
- ✅ `execute.go` 旧 `keyXxxVersion` 常量 — 已删除
- ✅ domain 层 `SystemSettings().Increment("catalog.xxx")` — 已替换为 `SyncVersions().Increment`
- ✅ 路由 `/sync/catalog/discounts` — 已注册
- ✅ `/sync/versions` — 已移入 sync-token-required 组
- ✅ `FetchVersions` client — 已带 auth

`system_settings` 表保留用于 Local 部署的 setup 配置：
- `catalog_sync_token`
- `setup_company_id`
- `setup_admin_email`
- `platform_channel_id`
- `saas_platform_key`
- `setup_idempotency_key`

---

## 8. 边界情况

1. **新建公司** — `sync_versions` 无行 → `Get` 返回 0 → Local 也是 0 → 不触发 sync（正确，无数据）
2. **首次 discount 配置** — Upsert version=1 → Local 检测到差异 → sync
3. **并发 bump** — `ON CONFLICT DO UPDATE SET version = version + 1` 原子安全
4. **公司删除** — 无 FK cascade，由删除公司的 handler/service 负责清理 `sync_versions` 行
5. **Local 首次同步** — local version=0（无行），remote > 0 → 触发 sync
6. **GlobalSyncVersion 行** — sentinel UUID，永远存在，不会被误删

---

## 9. 不做的事

- ❌ 不做 push-based 通知（保持轮询）
- ❌ 不做 optional auth fallback（项目未上线，直接 require auth）
- ❌ 不删 `system_settings` 表（仍有 Local setup 用途）
- ❌ 不加 FK（sentinel UUID 无对应 companies 行）

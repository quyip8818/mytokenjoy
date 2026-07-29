# 合同特殊价格

> per-company per-model 合同定价。入账时覆盖 quota，wallet-sync 自动保持 NewAPI wallet 精确。

---

## 1. 价格解析优先级

```
入账时确定 entry.QuotaAmount：
  model_contract_pricing[company_id + model_type]  → 合同价
  ↓ 没有
  raw.Quota（NewAPI 按全局 ModelRatio 算的）         → 全局价
```

---

## 2. 存储

```sql
CREATE TABLE model_contract_pricing (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id    UUID NOT NULL,
    model_type    VARCHAR(128) NOT NULL,
    input_price   NUMERIC(12,6) NOT NULL,
    output_price  NUMERIC(12,6) NOT NULL,
    note          TEXT DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, model_type)
);
```

---

## 3. 入账覆盖

位置：`IngestRaw` → `BuildCallSettledEntry` 之后、`ConsumeLotsLocked` 之前。

```go
func applyContractPricing(entry types.UsageLedgerEntry, pricing []store.ContractPricing, companyID uuid.UUID) types.UsageLedgerEntry {
    for _, cp := range pricing {
        if cp.CompanyID == companyID && cp.ModelType == entry.Model {
            ratio, compRatio := modelcatalog.RatioFromPrice(cp.InputPrice, cp.OutputPrice)
            entry.QuotaAmount = int64(math.Round(
                float64(entry.InputTokens)*ratio +
                    float64(entry.OutputTokens)*ratio*compRatio,
            ))
            entry.CallDetail.ContractPricing = true
            return entry
        }
    }
    return entry // 无合同价，透传 raw.Quota
}
```

合同价格在 `EntryBuildSnapshot` 中预加载：

```go
type EntryBuildSnapshot struct {
    Catalog         []types.ModelInfo
    OrgTree         []types.OrgNode
    ContractPricing []store.ContractPricing
}
```

---

## 4. Wallet 精确性

无需额外处理。wallet-sync 机制保证：

```
ConsumeLotsLocked(合同价 quota) → SetWalletRemainQuota(新余额)
→ post-commit: ManageUser("set_quota", 新余额) → NewAPI wallet = SOT
```

---

## 5. Local 同步

SaaS 暴露：
```
GET /api/platform/sync/catalog/contract-pricing?companyId=<uuid>
→ { "data": [{ "modelType": "deepseek-chat", "inputPrice": 1.5, "outputPrice": 3 }] }
```

Local catalogsync 拉取后写入本地 `model_contract_pricing` 表，入账走同样逻辑。

---

## 6. 管理 API

```
GET    /api/platform/companies/:companyId/pricing
PUT    /api/platform/companies/:companyId/pricing       { modelType, inputPrice, outputPrice, note? }
DELETE /api/platform/companies/:companyId/pricing/:modelType
```

校验：`inputPrice > 0`，`outputPrice > 0`，`inputPrice ≤ 全局 inputPrice`（推荐）。

Platform admin only。

---

## 7. 实施步骤

| # | 改动 |
|---|------|
| 1 | schema + store: `model_contract_pricing` CRUD |
| 2 | `EntryBuildSnapshot` 加载合同价格 |
| 3 | `applyContractPricing` 入账覆盖 |
| 4 | Platform admin CRUD API |
| 5 | SaaS catalog endpoint + Local catalogsync 拉取 |

# 合同特殊价格

> per-company per-model 合同定价。TokenJoy 入账时用合同价覆盖 quota，wallet-sync 机制自动保持 NewAPI wallet 精确。

---

## 1. 背景

- SaaS platform admin 需要给特定 company 的特定模型设置合同价格
- NewAPI `ModelRatio` 是全局的，无法区分 company
- 但 wallet-sync 机制（见 `docs/Backend-计费模式.md` §4.4）已经保证：**每次 TokenJoy wallet 变更后，post-commit override 到 NewAPI**
- 这意味着：只要 TokenJoy 侧 quota 计算正确，NewAPI wallet 自动保持精确

---

## 2. 为什么 wallet-sync 让这个问题变简单

```
当前 wallet-sync 流程：
  IngestRaw → ConsumeLotsLocked(entry.QuotaAmount) → SetWalletRemainQuota
  → post-commit: ManageUser("set_quota", 新余额)  // override 到 NewAPI

引入合同价后：
  IngestRaw → applyContractPricing → 修改 entry.QuotaAmount
  → ConsumeLotsLocked(修改后的 QuotaAmount) → SetWalletRemainQuota
  → post-commit: ManageUser("set_quota", 新余额)  // 自动正确！
```

**关键洞察：** wallet-sync 同步的是 `wallet_remain_quota` 的绝对值（SOT），不是 quota 差值。所以无论 `entry.QuotaAmount` 是全局价还是合同价，lot 扣减完后的 `wallet_remain_quota` 就是正确的值，override 到 NewAPI 也是正确的。

**不需要额外的对齐机制。** wallet-sync 已经解决了这个问题。

---

## 3. 方案

唯一改动：在 `BuildCallSettledEntry` 之后覆盖 `entry.QuotaAmount`。

```
NewAPI: quota = tokens × 全局 ModelRatio → 写入 logs 表
          ↓ webhook / reconcile
TokenJoy IngestRaw:
  raw.Quota = NewAPI 算的（全局价）
  entry = BuildCallSettledEntry(raw)        → entry.QuotaAmount = raw.Quota
  entry = applyContractPricing(entry, ...)  → entry.QuotaAmount = 合同价重算 ← 改这里
  ConsumeLotsLocked(entry.QuotaAmount)      → lot 按合同价扣减
  SetWalletRemainQuota(新余额)
  → post-commit: set_quota(新余额) → NewAPI wallet 精确 ✓
```

下游全部不感知。NewAPI wallet 自动精确。

---

## 4. 存储

```sql
CREATE TABLE model_contract_pricing (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id    UUID NOT NULL,
    model_type    VARCHAR(128) NOT NULL,
    input_price   NUMERIC(12,6) NOT NULL,   -- 元/1M tokens
    output_price  NUMERIC(12,6) NOT NULL,
    note          TEXT DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, model_type)
);
```

---

## 5. 入账覆盖逻辑

```go
// internal/domain/usage/contract_pricing.go

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
    return entry
}
```

加载位置：`EntryBuildSnapshot` 中预加载（和 Catalog、OrgTree 一起）：

```go
type EntryBuildSnapshot struct {
    Catalog         []types.ModelInfo
    OrgTree         []types.OrgNode
    ContractPricing []store.ContractPricing  // 新增
}
```

---

## 6. 管理 API

```
GET    /api/platform/companies/:companyId/pricing
PUT    /api/platform/companies/:companyId/pricing       { modelType, inputPrice, outputPrice, note? }
DELETE /api/platform/companies/:companyId/pricing/:modelType
```

### 输入校验

```go
if body.InputPrice <= 0 || body.OutputPrice <= 0 {
    return error("price must be positive")
}
// 可选：限制合同价 ≤ 全局价（防止透支风险）
if body.InputPrice > globalInputPrice {
    return error("contract price cannot exceed global price")
}
```

### 权限

Platform admin only（放在 platform 路由组下，和 SetModelPricing 同级）。

---

## 7. Local 同步

SaaS 暴露：
```
GET /api/platform/sync/catalog/contract-pricing?companyId=<uuid>
→ { "data": [{ "modelType": "deepseek-chat", "inputPrice": 1.5, "outputPrice": 3 }] }
```

Local catalogsync 拉取后写入本地表。Local 入账时走同样的 `applyContractPricing` 逻辑。

---

## 8. call_detail 审计

```go
type UsageCallDetail struct {
    // ... existing
    ModelRatio      float64 `json:"modelRatio,omitempty"`      // NewAPI 实际用的全局 ratio
    CompletionRatio float64 `json:"completionRatio,omitempty"` // NewAPI 实际用的全局 ratio
    ContractPricing bool    `json:"contractPricing,omitempty"` // 是否使用了合同价
}
```

对账：`raw.Quota`（全局价）vs `entry.QuotaAmount`（合同价），差值 = 折扣金额。

---

## 9. 与 model_changelog 联动

```
event_type: "contract_pricing_set"
payload: { "companyId": "...", "modelType": "...", "inputPrice": X, "outputPrice": Y }

event_type: "contract_pricing_removed"
payload: { "companyId": "...", "modelType": "..." }
```

---

## 10. 实施步骤

| # | 改动 | 优先级 |
|---|------|--------|
| 1 | schema: `model_contract_pricing` 表 | P0 |
| 2 | store: `ContractPricingRepository` CRUD | P0 |
| 3 | `EntryBuildSnapshot` 加载合同价格 | P0 |
| 4 | `applyContractPricing` 入账覆盖逻辑 | P0 |
| 5 | `UsageCallDetail` 加 `ContractPricing` 标记 | P0 |
| 6 | Platform admin CRUD API + 输入校验 | P0 |
| 7 | SaaS catalog endpoint 暴露合同价格 | P1 |
| 8 | Local catalogsync 拉取合同价格 | P1 |
| 9 | 前端 platform admin 管理页面 | P1 |
| 10 | 前端用户侧展示合同价 | P2 |

---

## 11. 注意事项

- **零 NewAPI patch** — 不动 NewAPI 任何代码
- **wallet 自动精确** — wallet-sync 机制保证每次入账后 override 到 NewAPI，合同价导致的 lot 扣减差异自动反映到 NewAPI wallet
- **math.Round** — 避免浮点截断导致累积偏差
- **inputPrice > 0 校验** — 防止设为 0 免费使用
- **合同价 ≤ 全局价校验**（推荐） — 防止 NewAPI 在 override 前的窗口内漏放请求
- 即时生效，不追溯历史 ledger
- 删除合同价 = 恢复全局价
- Platform admin only 权限

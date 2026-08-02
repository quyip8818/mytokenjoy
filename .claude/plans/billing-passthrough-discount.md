# 计费架构重构：Passthrough + Discount

## 设计原则

**TJ 是企业管控层，不是计费引擎。**

- 定价：由 NewAPI 负责（支持 per-token / per-request / billingexpr / audio / cache / video 等所有模式）
- TJ 职责：passthrough 上游 quota + 企业级折扣 + 钱包/lot 扣费 + 预算管控 + 审计

---

## 系统拓扑

```
┌─────────────────────────────────────────────────────────────────────┐
│                     SaaS Platform（TJ 总部）                         │
│                                                                     │
│  models ───┐                                                        │
│  prices ───┼── /api/platform/sync/catalog/* ──────────────────┐     │
│  discounts─┤                                                  │     │
│  currencies┤                                                  │     │
│  lots ─────┘                                                  │     │
└───────────────────────────────────────────────────────────────┼─────┘
                                                                │
                        CatalogSync (version-gated pull)         │
                                                                │
┌───────────────────────────────────────────────────────────────▼─────┐
│                    Local Instance（客户私有化）                       │
│                                                                     │
│  ┌─────────────┐         ┌───────────────┐                         │
│  │   TJ App    │◄─ingest─│    NewAPI      │◄──── 上游 LLM Provider  │
│  │             │         │  (API Gateway)  │                         │
│  │ • discount  │─ratio──▶│  • ModelRatio   │                         │
│  │ • lot/wallet│         │  • logs 表      │                         │
│  │ • budget    │         └───────────────┘                         │
│  │ • audit     │                                                    │
│  └─────────────┘                                                    │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 核心计费公式

```
finalQuota = raw.Quota × discount
```

- `raw.Quota`：NewAPI 按 ModelRatio + 实际用量（含 cache/audio/video 等所有模式）算出的 quota
- `discount`：TJ 侧按公司+模型查出的折扣系数（默认 1.0 = 原价透传）

### 前提条件：ModelRatio 与 TJ 全局价一致

passthrough 的正确性依赖于：NewAPI 的 ModelRatio 和 TJ 的全局定价同步。

详见 → [model-ratio-sync.md](./model-ratio-sync.md)

**结论**：已有 4 条 best-effort 推送路径，但缺少全量兜底定时任务。**Phase 0 必须先修复**。

---

## 数据模型

### 1. models 表（改造）

```sql
-- 新增两列，CatalogSync 写入
ALTER TABLE models ADD COLUMN input_price  NUMERIC(12,6) NOT NULL DEFAULT 0;
ALTER TABLE models ADD COLUMN output_price NUMERIC(12,6) NOT NULL DEFAULT 0;
```

| 字段 | 用途 |
|------|------|
| `input_price` | 展示用（前端"模型列表"页面显示单价），CatalogSync 从 SaaS 同步 |
| `output_price` | 同上 |
| 其他字段 | 不变 |

**这两个字段不参与计费计算**，纯展示 + 推到 NewAPI 作为 gateway ratio。

### 2. model_discount 表（新建）

企业级折扣。Append-only 支持历史追溯。

```sql
CREATE TABLE model_discount (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    model_type     TEXT NOT NULL,
    discount       NUMERIC(5,4) NOT NULL DEFAULT 1.0,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note           TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_discount_current
    ON model_discount (company_id, model_type, effective_from DESC);
```

语义：
- `discount = 1.0`：原价
- `discount = 0.8`：八折（客户优惠）
- `discount = 1.2`：加价 20%（TJ 毛利空间）
- 无记录 = 1.0（隐式原价）

> **约束策略**：DB 不加 CHECK。前端输入时 `< 0.5 || > 2.0` 弹二次确认弹窗。

### 3. model_history（暂不新建）

当前 `model_pricing` 已经是 append-only 时间线，天然带历史。
新方案中 `models.input_price / output_price` 变为 mutable 列，价格变更记录通过以下方式覆盖：

- CatalogSync pricing 写入时，如果价格变化，写一条 slog.Info（审计日志）
- 未来如需前端展示价格历史，再加 `model_history` 表（YAGNI）

> ponytail: model_history 的 ROI 目前不够。事件类型杂（created/price_changed/activated/deactivated），snapshot 是非结构化 JSONB，前端渲染成本高。升级路径：需要时再加，只记 price_changed 事件即可。

### 4. model_pricing 表（删除）

不再需要。全局价归 `models` 表，合约价归 `model_discount` 表。

---

## 关键流程

### CatalogSync（SaaS → Local）

```go
func (e *Executor) Execute(ctx context.Context) error {
    // 1. Models sync（不变）
    // 2. Pricing sync（改造）
    // 3. Discounts sync（新增）
    // 4. Currencies sync（不变）
    // 5. Wallet lots sync（不变）
}
```

#### Pricing sync 改造

SaaS 端 `/api/platform/sync/catalog/pricing` 响应改为：

```json
{
  "version": 5,
  "data": [
    {"modelType": "gpt-4o", "inputPrice": 2.5, "outputPrice": 10.0},
    {"modelType": "claude-3.5-sonnet", "inputPrice": 3.0, "outputPrice": 15.0}
  ]
}
```

- 只返回全局价（不再返回 `isContract`）
- Local 收到后：
  1. 更新 `models.input_price / output_price`
  2. 推到 NewAPI (`UpsertModelRatio`)

#### Discount sync（新增通道）

SaaS 端新增 `/api/platform/sync/catalog/discounts`：

```json
{
  "version": 3,
  "data": [
    {"modelType": "gpt-4o", "discount": 0.8},
    {"modelType": "claude-3.5-sonnet", "discount": 0.85},
    {"modelType": "*", "discount": 0.9}
  ]
}
```

Local 收到后写入 `model_discount` 表（append-only，`effective_from = now()`）。

`system_settings` 增加 `catalog.discounts_version` 做增量判断。

### Ingest 流程（改造后）

```go
func (s *IngestService) IngestRaw(ctx context.Context, raw store.RawConsumeLog, source string) error {
    // ... mapping, context 解析（不变）...

    snap, _ := LoadEntryBuildSnapshot(ctx, s.store, s.cfg.TokenJoyCompanyID)
    buildInput, _ := LoadEntryBuildInput(ctx, s.store, mapping, raw, source, snap)
    entry, _ := BuildCallSettledEntry(buildInput)

    // NEW: passthrough + discount（替代旧的 ApplyTJPricing）
    entry = ApplyDiscount(entry, snap.Discounts)

    // ... 后续 lot 消费、budget、ledger 写入（不变）...
}
```

#### ApplyDiscount 实现

```go
func ApplyDiscount(entry types.UsageLedgerEntry, discounts []store.ModelDiscountRow) types.UsageLedgerEntry {
    d := resolveDiscount(discounts, entry.Model)
    if d == 1.0 {
        return entry
    }
    entry.QuotaAmount = int64(math.Ceil(float64(entry.QuotaAmount) * d))
    entry.CallDetail.Discount = d
    entry.CallDetail.ContractPricing = true
    return entry
}

// resolveDiscount: 精确匹配 > 通配 "*" > 默认 1.0
func resolveDiscount(discounts []store.ModelDiscountRow, modelType string) float64 {
    var wildcard float64 = 1.0
    for _, d := range discounts {
        if d.ModelType == modelType {
            return d.Discount
        }
        if d.ModelType == "*" {
            wildcard = d.Discount
        }
    }
    return wildcard
}
```

### EntryBuildSnapshot 简化

```go
type EntryBuildSnapshot struct {
    Catalog      []types.ModelInfo
    OrgTree      []types.OrgNode
    Discounts    []store.ModelDiscountRow  // NEW: 替代 CompanyPricing + GlobalPricing
    QuotaPerUnit int64
}
```

`LoadEntryBuildSnapshot` 只需加载 `model_discount` 表当前生效行（`WHERE company_id = ?`），查询更简单。

---

## QPU 一致性保证

| 层 | QPU 来源 | 值 |
|---|---|---|
| NewAPI | TJ 推送 `UpsertModelRatio`（基于 models 表的展示价转换） | `QuotaPerUnit` = currencies 表 |
| TJ lot/wallet | currencies 表（CatalogSync 同步） | 相同 |

**raw.Quota 和 TJ lot 用同一套单位**，无需换算。CatalogSync currencies 保证两侧一致。

---

## 实现状态

已完成。项目未上线，无旧数据，直接重建 schema 即可。

- ✅ `model_discount` 表已创建
- ✅ `models.input_price / output_price` 列已添加
- ✅ `model_pricing` 表已删除
- ✅ `ApplyDiscount` 已替代 `ApplyTJPricing`
- ✅ CatalogSync 适配（pricing 去 IsContract + 新增 discount 通道）
- ✅ `FullSyncToNewAPI` 注册为 River PeriodicJob（5min）
- ✅ HTTP API：全局价 `GET/PUT /pricing`，折扣 `GET/PUT /companies/{id}/discounts`
- ✅ 旧 `CalcQuota`、`pricing_override.go`、`model_pricing_repo.go` 已清理

---

## 前端影响（待实现）

| 页面 | 变更 |
|------|------|
| 模型列表 | 价格读 `models.input_price / output_price` |
| 合约价设置 | 改为"输入折扣倍率"。`< 0.5 \|\| > 2.0` 弹二次确认 |
| 平台定价管理 | 简化：只管全局价 |

---

## 风险与缓解

| 风险 | 缓解 |
|------|------|
| NewAPI ratio 和 TJ 展示价短暂不一致 | FullSyncToNewAPI 定时 5min 全量对齐兜底 |
| discount 配错（如误设为 0.01） | 前端 `<0.5 \|\| >2.0` 弹二次确认 + 操作审计 |
| NewAPI 侧 quota 计算有 bug | TJ 在 ledger call_detail 中记录 raw.Quota 原始值，支持事后对账 |
| 通配 `*` discount 优先级冲突 | 精确匹配 > 通配，规则简单明确 |

---

## 最终架构收益

1. **正确性**：任何计费模式（token/次/秒/cache/audio/billingexpr）都能稳健计费
2. **简单性**：核心计费逻辑从 ~100 行 → 10 行（一次乘法）
3. **可维护性**：NewAPI 上新计费模式不需要 TJ 改代码
4. **职责清晰**：NewAPI = 定价引擎，TJ = 折扣 + 管控 + 分账
5. **可审计**：model_discount append-only 支持合规追溯，ledger 记录 raw.Quota + discount 完整链路

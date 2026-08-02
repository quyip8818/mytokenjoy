# NewAPI as 定价 SOT：消除价格同步

## 动机

Passthrough 架构下，`finalQuota = raw.Quota × discount`，raw.Quota 由 NewAPI 的 ModelRatio 计算。
既然计费完全依赖 NewAPI 的 ratio，那 ratio 本身就是 SOT——TJ 不需要管理价格、不需要缓存价格、不需要同步价格。

---

## 架构

```
NewAPI ModelRatio/CompletionRatio (SOT)
    │
    │  logs.quota = f(tokens, ratio)
    ▼
TJ passthrough: finalQuota = raw.Quota × discount
```

展示价格：前端需要时，Backend 实时调 NewAPI `ListModelPricing()` → `PriceFromRatio` 转换返回。
不缓存、不存 DB。NewAPI 同机部署，延迟 <1ms，模型数 <100。

---

## 删除清单

| 删除 | 理由 |
|------|------|
| `models.input_price / output_price` 列 | 不需要缓存，实时从 NewAPI 读 |
| `ModelsRepository.UpdatePrice()` | 无写入需求 |
| `domain/pricing/service.go` 整个文件 | 不再管理价格 |
| `FullSyncToNewAPI()` + `PricingFullSyncArgs` + `PricingFullSyncWorker` | 不需要推送 |
| `periodic/jobs.go` 中 pricing sync job | 不需要定时同步 |
| River Deps 中 `PricingSvc` 字段 | pricing 域消除 |
| `compose_domain_wire.go` 中 `wirePricing` | 同上 |
| `http/deps` 中 `PricingSvc` 字段 | 同上 |
| CatalogSync `syncPricing` 中写 models 表的逻辑 | 改为只 push ratio 到 NewAPI |
| `.claude/plans/model-ratio-sync.md` | 问题不存在了 |

---

## 保留

| 保留 | 用途 |
|------|------|
| `adminPort.ListModelPricing()` | 展示：读 NewAPI ratio → 转换为展示价 |
| `adminPort.UpsertModelRatio()` | 改价代理：TJ UI 改价 → 直接写 NewAPI |
| `pkg/modelcatalog/pricing.go` PriceFromRatio/RatioFromPrice | 转换工具 |
| `model_discount` 表 + `ApplyDiscount` | 折扣逻辑（不变） |

---

## 需要修改（非删除）的模块

### 1. SaaS CatalogPricing endpoint（`platform/catalog_pricing.go`）

当前从 `PricingSvc.ListGlobalPricing()` 读 DB。删 DB 价格列后需改为：

```go
func (h *Handler) CatalogPricing(w http.ResponseWriter, r *http.Request) {
    // 实时读 NewAPI ratio → 转换为展示价
    ratios, err := h.p.AdminPort.ListModelPricing(r.Context())
    // ... PriceFromRatio 转换 → 返回
}
```

### 2. models/service.go CreateModel/UpdateModel

当前写 `model.InputPrice/OutputPrice` 到 DB 并 best-effort push NewAPI。改为：
- 删除 DB 写入（列已不存在）
- 保留 `UpsertModelRatio` push（用户设的价格直接写 NewAPI SOT）
- `CreateModelInput`/`UpdateModelInput` 的价格字段保留（UI 需要传值）

### 3. platform/models handler CreateModel

当前调 `h.p.PricingSvc.SetGlobalPrice()`。改为直接调 `h.p.AdminPort.UpsertModelRatio()`。

### 4. pricing_version bump 机制

`SetGlobalPrice` 每次改价 bump `catalog.pricing_version`，CatalogSync 客户端靠版本号判断是否需要重拉价格。

删 pricing service 后需要新机制：
- **方案 A**：在 `UpsertModelRatio` 代理方法或新 handler 写 NewAPI 之后 bump version（推荐）
- **方案 B**：Local 改为 always-sync（每次 CatalogSync 都读价格），不依赖 version

推荐 A：在 platform/models handler 和 admin pricing handler 写 NewAPI 后，顺带 bump `catalog.pricing_version`。

### 5. ListModelsWithPricing 实现

`models/handler.go` 调 `ListModelsWithPricing`。改为在 models service 内实时 merge NewAPI 价格：

```go
func (s *service) ListModelsWithPricing(ctx context.Context) ([]types.ModelInfo, error) {
    models, err := s.ListModels(ctx)
    if err != nil { return nil, err }
    // s.client 即 adminPort（已持有）
    ratios, _ := s.client.ListModelPricing(ctx)
    priceMap := buildPriceMap(ratios) // map[modelType]{input, output}
    for i := range models {
        if p, ok := priceMap[models[i].Type]; ok {
            models[i].InputPrice = p.InputPrice
            models[i].OutputPrice = p.OutputPrice
        }
    }
    return models, nil
}
```

### 6. 自定义模型 scope 确认

自定义模型也 `UpsertModelRatio` 到 NewAPI。需确认：
- NewAPI ratio key 是否 globally unique（当前是 modelType 作 key）
- 多租户场景下不同 company 自定义同名 modelType 是否冲突

结论：当前 TJ 架构只有 global（TokenJoy company）模型走 pricing path，自定义模型跟随其 modelType 唯一性约束（`UNIQUE(company_id, provider, type)`），不会 cross-tenant 冲突。**前提：每个部署只有一个 NewAPI 实例。** 如果未来多租户共享 NewAPI，需要加 namespace。

---

## 读写路径

### 展示价格（读）

```
GET /api/models (前端) 或 GET /api/platform/pricing
    → handler 调 adminPort.ListModelPricing()
    → 得到 []ModelPricing{ModelName, ModelRatio, CompletionRatio}
    → PriceFromRatio(ratio, completionRatio) 转换为 (inputPrice, outputPrice)
    → 返回前端
```

不经过 DB。每次请求实时读 NewAPI option store。

### 改全局价（写）

```
PUT /api/platform/pricing {modelType, inputPrice, outputPrice}
    → RatioFromPrice(inputPrice, outputPrice) 转换为 (modelRatio, completionRatio)
    → adminPort.UpsertModelRatio(modelType, inputPrice, outputPrice)
    → 写入 NewAPI option store（SOT）
    → bump catalog.pricing_version（供 CatalogSync 客户端检测变更）
    → 返回 OK
```

TJ 只做代理转发，不存任何价格数据。

### CatalogSync（SaaS → Local 私有化）

```
SaaS /sync/catalog/pricing 返回 [{modelType, inputPrice, outputPrice}]
    （SaaS 侧也实时读 NewAPI ListModelPricing + PriceFromRatio 转换）
    → CatalogSync 遍历
    → adminPort.UpsertModelRatio(modelType, inputPrice, outputPrice)
    → 直接写 Local NewAPI（SOT）
```

不写 TJ models 表。

### 计费（不变）

```
NewAPI logs.quota → TJ IngestRaw → ApplyDiscount(quota, discounts) → lot consume
```

---

## 前端模型列表页

当前前端通过 `GET /api/models` 拿模型列表 + 价格。新方案：

- 模型目录（name/type/provider/active）：仍从 TJ models 表读
- 价格：models service 内部调 `ListModelPricing()` merge 到返回结果

`ModelInfo` 的 `InputPrice/OutputPrice` 字段保留在 struct 上（JSON 响应用），但不存 DB。

---

## models 表 schema

```sql
CREATE TABLE models (
    model_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,
    type         TEXT NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    endpoint     TEXT,
    api_key      TEXT,
    endpoint_model_name TEXT,
    max_context  INT NOT NULL DEFAULT 0,
    max_tokens   INT NOT NULL DEFAULT 0,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    capabilities TEXT[] NOT NULL DEFAULT '{}',
    source       TEXT NOT NULL DEFAULT 'manual',
    catalog_synced_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, provider, type)
);
```

**无 price 列。** 价格实时从 NewAPI 读。

---

## 实现步骤

1. Schema: 删除 `models.input_price / output_price` 列
2. 删除 `pricing/service.go`、`kinds_pricing_sync.go`、`workers/pricing_sync.go`
3. 删除 River Deps 中 PricingSvc、periodic jobs 中 pricing sync
4. 删除 `compose_domain_wire.go` 中 `wirePricing`、`http/deps` 中 `PricingSvc`
5. 删除 `ModelsRepository.UpdatePrice()` / `ModelsByCompany()`（如果无其他使用者）
6. **SaaS CatalogPricing endpoint**：改为实时读 NewAPI `ListModelPricing()` + `PriceFromRatio`
7. **models/service CreateModel/UpdateModel**：删 DB 价格写入，保留 `UpsertModelRatio` push
8. **platform/models handler**：`PricingSvc.SetGlobalPrice` → `adminPort.UpsertModelRatio` + bump version
9. **pricing_version bump**：在所有改价路径末尾 bump `catalog.pricing_version`
10. **ListModelsWithPricing**：改为 models service 内实时读 NewAPI merge
11. CatalogSync `syncPricing`：只 push ratio 到 NewAPI，不写 models 表
12. 删除 `.claude/plans/model-ratio-sync.md`
13. 确认 go build 通过

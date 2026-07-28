# 模型定价读写路径

## 存储位置

模型价格 **不存在 models 表**，只存在 NewAPI 的 `options` 表中：

| key | 格式 | 示例 |
|-----|------|------|
| `ModelRatio` | JSON map: `{modelType: ratio}` | `{"gpt-4o": 1.25, "deepseek-v4-pro": 7.5}` |
| `CompletionRatio` | JSON map: `{modelType: ratio}` | `{"deepseek-v4-pro": 4, "gpt-4o": 2}` |

## 价格换算公式

```
modelRatio      = inputPrice / 2
completionRatio = outputPrice / inputPrice

反向：
inputPrice  = modelRatio * 2
outputPrice = modelRatio * completionRatio * 2
```

代码位置：`integration/newapi/option.go` → `RatioFromPrice()`，`pkg/modelcatalog/pricing.go` → `PriceFromRatio()`

## 写入路径

```
platform admin 设定价格
  │
  ├─ PUT /api/platform/models/:id/pricing  { inputPrice, outputPrice }
  │     → handler/platform/models.go SetModelPricing
  │
  └─ adminport.UpsertModelRatio(modelType, inputPrice, outputPrice)
       → integration/newapi/option.go UpsertModelRatio
       │
       ├─ GET  NewAPI /api/option/       → 读取现有 ModelRatio + CompletionRatio map
       ├─ 计算 modelRatio, completionRatio
       ├─ PUT  NewAPI /api/option/ key=ModelRatio     value=更新后的 JSON map
       └─ PUT  NewAPI /api/option/ key=CompletionRatio value=更新后的 JSON map
```

同样的写入路径也用于：
- `POST /api/platform/models`（创建模型时附带价格）
- `PUT /api/models/:id`（客户自定义模型更新价格）
- `catalogsync` worker（local 从 SaaS 同步价格）

## 读取路径

```
任何需要展示价格的 API
  │
  ├─ GET /api/platform/models         → handler/platform/models.go ListModels
  ├─ GET /api/models                  → handler/models/handler.go List
  ├─ GET /api/platform/sync/catalog/models → handler/platform/models.go CatalogModels
  │
  └─ adminport.ListModelPricing()
       → integration/newapi/pricing.go ListModelPricing
       │
       ├─ GET  NewAPI /api/option/     → 读取 ModelRatio + CompletionRatio JSON map
       ├─ 解析 JSON map
       └─ 返回 []ModelPricing{ ModelName, ModelRatio, CompletionRatio }

展示层：
  inputPrice, outputPrice = PriceFromRatio(modelRatio, completionRatio)
```

## 关键约束

1. **读写同源**：`UpsertModelRatio` 和 `ListModelPricing` 都通过 NewAPI `/api/option/` 端点操作同一组 key。
2. **models 表不存价格**：价格仅在 NewAPI option 中，models 表只存目录信息。
3. **best-effort 读取**：`ListModelPricing` 失败时返回空（价格显示为 0），不阻塞模型列表展示。
4. **NewAPI auth**：需要 `Authorization: Bearer <token>` + `New-Api-User: <adminUserID>` 两个 header。

## 文件索引

| 职责 | 文件 |
|------|------|
| 写入定价 | `internal/integration/newapi/option.go` → `UpsertModelRatio` |
| 读取定价 | `internal/integration/newapi/pricing.go` → `ListModelPricing` |
| 价格换算 | `internal/pkg/modelcatalog/pricing.go` → `PriceFromRatio` / `RatioFromPrice` |
| Platform handler | `internal/http/handler/platform/models.go` → `SetModelPricing` / `fetchPriceMap` |
| 客户侧 handler | `internal/domain/models/service.go` → `ListModelsWithPricing` |
| adminport 接口 | `internal/domain/adminport/port.go` → `UpsertModelRatio` / `ListModelPricing` |

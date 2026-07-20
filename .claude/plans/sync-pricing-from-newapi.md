# 从 NewAPI 同步模型价格到 Backend 模型目录

## 目标

Backend 模型目录的 `InputPrice`/`OutputPrice` 从 NewAPI 的定价数据自动同步，而不再由 seed 硬编码。

## 转换公式（已验证）

```
InputPrice (display, 元/1M tokens) = model_ratio * 2
OutputPrice (display, 元/1M tokens) = model_ratio * completion_ratio * 2
内部存储 points = display × DefaultPointsPerUnit (1000)
```

示例：test-model ratio=0.075, completion_ratio=4 → input=0.15, output=0.60 ✓

## 匹配策略

精确匹配：Backend `models.type` == NewAPI `/api/pricing` 返回的 `model_name`。

## 触发方式

1. **启动自动同步**：Backend 启动时执行一次
2. **手动 API**：管理员通过 API 按钮触发

## 实现步骤

### 1. AdminPort 扩展：添加 `ListModelPricing` 方法

**文件**: `internal/domain/adminport/port.go` + `types.go`

```go
// port.go - Port interface 新增方法
ListModelPricing(ctx context.Context) ([]ModelPricing, error)

// types.go - 新增类型
type ModelPricing struct {
    ModelName       string
    ModelRatio      float64
    CompletionRatio float64
}
```

### 2. AdminClient 实现：调用 NewAPI `/api/pricing`

**文件**: `internal/integration/newapi/pricing.go`（新文件）

- 调用 `GET /api/pricing`（无需 auth，公开接口）
- 解析返回的 `data[]` 数组，提取 `model_name`, `model_ratio`, `completion_ratio`
- 返回 `[]adminport.ModelPricing`

### 3. Models Domain：添加 `SyncPricingFromUpstream` 方法

**文件**: `internal/domain/models/service.go`

- Service interface 新增 `SyncPricingFromUpstream(ctx context.Context) (int, error)` 方法
- 实现逻辑：
  1. 调用 `adminPort.ListModelPricing()` 获取 NewAPI 全部定价
  2. 调用 `store.Models().ListAll()` 获取 Backend 所有模型
  3. 对于每个 Backend 模型，按 `type` 精确匹配 NewAPI pricing
  4. 如果找到且价格不同，计算新价格并 `UpdateModel`
  5. 返回更新数量

转换代码放 `internal/pkg/newapiunits/pricing.go`（新文件）：
```go
func PriceFromRatio(modelRatio, completionRatio float64) (inputPoints, outputPoints float64) {
    inputDisplay := modelRatio * 2
    outputDisplay := modelRatio * completionRatio * 2
    return inputDisplay * common.DefaultPointsPerUnit, outputDisplay * common.DefaultPointsPerUnit
}
```

### 4. HTTP Handler：`POST /api/models/sync-pricing`

**文件**: `internal/http/handler/models/handler.go`

- 需要平台管理员权限
- 调用 `service.SyncPricingFromUpstream(ctx)`
- 返回 `{ "updated": N }`

### 5. 启动时自动同步

**文件**: `internal/app/` 相关启动代码

- 在 Backend 启动完成后（NewAPI 可达时），异步调用一次 `SyncPricingFromUpstream`
- 仅在 `NEW_API_ENABLED=true` 时执行
- 失败仅 log warning，不阻塞启动

### 6. 前端：模型管理页增加"同步价格"按钮

**文件**: `apps/frontend/src/features/models/`

- 模型列表页增加 "同步价格" 按钮
- 调用 `POST /api/models/sync-pricing`
- 显示成功后更新了多少个模型的价格

## 影响范围

- `seed/snapshot/models.go` 中的价格仅作为 fallback 初始值，同步后会被覆盖
- 已有的 `CostFromQuota` / `ToNewAPIUnits` 逻辑不变，它们读 DB 中的 price
- 不影响已有的 ProviderKey → Channel 单向推送逻辑

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `internal/domain/adminport/port.go` | 接口新增方法 |
| `internal/domain/adminport/types.go` | 新增 ModelPricing 类型 |
| `internal/integration/newapi/pricing.go` | 新文件：调用 /api/pricing |
| `internal/integration/newapi/client.go` | AdminClient 接口新增方法 |
| `internal/pkg/newapiunits/pricing.go` | 新文件：ratio→points 转换 |
| `internal/domain/models/service.go` | Service 新增 SyncPricingFromUpstream |
| `internal/http/handler/models/handler.go` | 新增 POST /api/models/sync-pricing |
| `internal/app/` | 启动时调用 sync |
| `apps/frontend/src/features/models/` | 同步按钮 UI |
| 测试文件 | 对应的单元测试 |

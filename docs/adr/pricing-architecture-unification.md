# ADR: 定价架构统一

状态：已实现

## 决策

所有模型定价（内置 + 自定义）统一由本地 NewAPI 管理。TJ models 表不存价格。

```
官方管理平台
    ↓ pricing sync worker（feature flag 控制）
本地 NewAPI (model_ratio / completion_ratio)
    ↓ GET /api/pricing
TJ models.Service.ListModelsWithPricing()
    ↓
前端展示
```

## 架构

### 价格唯一 SOT：本地 NewAPI

| 模型类型 | 价格写入 | 价格读取 |
|---------|---------|---------|
| 内置模型 | pricing sync worker → NewAPI option | `ListModelsWithPricing` → NewAPI |
| 自定义模型 | `CreateModel`/`UpdateModel` → `UpsertModelRatio` → NewAPI option | 同上 |

### models 表职责（纯目录，不含价格）

- 模型 UUID（外键引用：org_nodes、model_allowlist）
- provider / type / name / description
- endpoint / apiKey / endpointModelName（自定义模型连接信息）
- max_context / max_tokens / enabled / capabilities

### Service 层

```go
type Service interface {
    ListModelsWithPricing(ctx)  // DB 目录 + NewAPI 价格 join，给前端
    ListModels(ctx)             // 纯 DB，给路由/precheck/ingest 热路径
    CreateModel(ctx, input)     // 写 DB + 写 NewAPI ratio
    UpdateModel(ctx, id, input) // 写 DB + 价格变更时写 NewAPI ratio
    // ... 其他不变
}
```

## 已实现的改动

### 后端

| 文件 | 变更 |
|------|------|
| `store/postgres/schema.sql` | 删除 `input_price`/`output_price` 列 |
| `store/postgres/models_repo_*.go` | Insert/Update/Scan 去除价格字段 |
| `domain/types/models.go` | `InputPrice`/`OutputPrice` 保留（JSON API 响应用，不映射 DB） |
| `domain/models/service.go` | 删除 `SyncPricingFromUpstream`；新增 `ListModelsWithPricing`；Create/Update 写 NewAPI |
| `domain/adminport/port.go` | 新增 `UpdateOption`、`UpsertModelRatio` |
| `integration/newapi/option.go` | 实现 `UpdateOption`、`UpsertModelRatio`（read-modify-write ratio map） |
| `integration/platform/pricing.go` | 新增平台 HTTP client `GetLatestPricing` |
| `pkg/newapiunits/pricing.go` | 新增 `RatioFromPrice`（反向转换） |
| `worker/pricingsync/worker.go` | 新增 pricing sync worker（定时从平台拉价格写本地 NewAPI） |
| `config/config.go` | 新增 `PLATFORM_PRICING_SYNC_ENABLED/URL/KEY/INTERVAL_SEC` |
| `app/app.go` | 删除 startup goroutine；注册 pricing sync worker（feature flag） |
| `handler/models/handler.go` | List 改调 `ListModelsWithPricing`；删除 `/sync-pricing` 路由 |
| `seed/snapshot/models.go` | 去除价格字段 |

### 前端

| 文件 | 变更 |
|------|------|
| `features/models/components/model-list-table.tsx` | 新增输入/输出价格列（¥/M tokens） |
| `features/workflow/workflows/model-edit.tsx` | 内置模型价格字段只读 + 提示文案 |

## 未实现（低优先级）

| 项目 | 说明 |
|------|------|
| `handler/pricing/handler.go` | GET /pricing/sync-status, POST /pricing/sync-now 管理端点。当前 worker 自动运行，暂不需要手动触发。需要时再加。 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PLATFORM_PRICING_SYNC_ENABLED` | `false` | 是否启动 pricing sync worker |
| `PLATFORM_PRICING_SYNC_URL` | — | 平台 API base URL |
| `PLATFORM_PRICING_SYNC_KEY` | — | 平台 API 认证 key（Bearer） |
| `PLATFORM_PRICING_SYNC_INTERVAL_SEC` | `600` | 同步间隔（秒） |

## 平台 API 契约（待平台实现）

```
GET /api/v1/pricing/latest
Authorization: Bearer {INSTANCE_API_KEY}

Response:
{
  "version": "2026-07-21T00:00:00Z",
  "model_ratio": "{\"deepseek-v4-pro\": 0.5, ...}",
  "completion_ratio": "{\"deepseek-v4-pro\": 2.5, ...}"
}
```

Worker 代码已就绪，`PLATFORM_PRICING_SYNC_ENABLED=true` 即可启用。

## 关键决策

| 决策 | 理由 |
|------|------|
| models 表不存价格 | NewAPI 是所有定价的唯一 SOT，避免双写不一致 |
| 自定义模型价格也写 NewAPI | 统一读取路径，`ListModelsWithPricing` 无需区分来源 |
| `UpsertModelRatio` 用 read-modify-write | NewAPI option 是全局 JSON map，只能整体覆盖 |
| Worker 通过 feature flag 控制 | 平台 API 未就绪，不阻塞其他功能 |
| `ListModelsWithPricing` best-effort | NewAPI 不可达时返回价格为 0，不阻塞模型列表 |
| 热路径用 `ListModels`（纯 DB） | 路由/precheck 不需要价格，不付 HTTP 开销 |

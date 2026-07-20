# 自定义模型表单重设计实现计划

## 字段映射

| 表单字段 | 后端字段 | 必填 | 状态 |
|----------|----------|------|------|
| 模型名称 | `type` | ✅ | 已有 |
| 模型显示名称 | `name` | — | 已有 |
| API Key | `apiKey` | — | **新增** |
| API endpoint URL | `baseUrl` | ✅ | 已有（rename from endpoint） |
| API endpoint中的模型名称 | `endpointModelName` | — | **新增** |
| Completion mode | `capabilities` | — | 已有(select: 对话/嵌入等) |
| 模型上下文长度 | `maxContext` | ✅ | 已有(CreateInput需添加) |
| 最大 token 上限 | `maxTokens` | — | **新增** |
| 输入单价 | `inputPrice` | — | 已有 |
| 输出单价 | `outputPrice` | — | 已有 |

## 改动范围

### 后端

1. **`internal/domain/types/models.go`**
   - `ModelInfo` 新增字段：`ApiKey`, `EndpointModelName`, `MaxTokens`
   - `CreateModelInput` 新增：`ApiKey`, `EndpointModelName`, `MaxContext`, `MaxTokens`, `Capabilities`
   - `UpdateModelInput` 新增：`ApiKey`, `EndpointModelName`, `MaxTokens`

2. **`internal/domain/models/service.go`**
   - `CreateModel` 使用新字段

3. **`internal/store/postgres/models_repo_crud.go`**
   - INSERT/UPDATE/SELECT 添加新列

4. **数据库 migration** — 新增列 `api_key`, `endpoint_model_name`, `max_tokens`

### 前端

1. **`api/types/models.ts`** — 更新 `ModelInfo`, `CreateModelInput`, `UpdateModelInput`
2. **`workflow/workflows/model-create.tsx`** — 重写表单按新设计
3. **`workflow/workflows/model-edit.tsx`** — 同步更新编辑表单

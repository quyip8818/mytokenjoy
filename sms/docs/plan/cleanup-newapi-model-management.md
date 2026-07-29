# SMS 清理：移除 NewAPI 集成与模型管理

## 背景

模型管理（模型目录 + 定价 + 模型同步）已完整迁移至 apps SaaS（platform admin），见 `apps/docs/platform-model-management.md`。

SMS 的定位回归为**纯粹的供应商管理系统**（合同、采购、评估），不再承担：
- 维护面向客户的模型目录
- 与 NewAPI 实例同步定价
- 对外提供 Catalog Sync API

## 要删除的内容

### 后端

| 类别 | 路径 | 说明 |
|------|------|------|
| domain | `internal/domain/model/` | 整个模型 CRUD service |
| domain | `internal/domain/newapisync/` | NewAPI 定价同步 service（含 pull） |
| domain | `internal/domain/sync/` | Catalog Sync API service（供 apps local 拉取） |
| handler | `internal/http/handler/model/` | `/api/models` CRUD handler |
| handler | `internal/http/handler/newapisync/` | `/api/newapi` handler（status/sync/pull） |
| handler | `internal/http/handler/sync/` | `/api/sync` handler（catalog endpoint） |
| integration | `internal/integration/newapi/` | NewAPI HTTP client + TokenStore |
| store | `internal/store/postgres/model.go` | models 表 CRUD |
| store | `internal/store/postgres/sync_pull.go` | channels/synced models upsert |
| store | `internal/store/postgres/sync_store.go` | catalog sync 读取 |
| store | `internal/store/postgres/oauth_client_store.go` | OAuth client（仅供 sync API 鉴权） |
| config | `config.go` 中 NewAPI 相关字段 | `NewAPIBaseURL`, `NewAPIAdminUserID`, `NewAPIDatabaseURL`, 及其辅助方法 |
| deps | `internal/http/deps/deps.go` | 移除 `NewAPISync`, `OAuth`, `Sync`, `Model` 字段 |
| app | `internal/app/app.go` | 移除 newapi/model/oauth/sync 的初始化逻辑 |
| middleware | `internal/http/middleware/` | 移除 `OAuthGuard`（仅 sync API 在用） |

### 前端

| 类别 | 路径 | 说明 |
|------|------|------|
| feature | `src/features/models/` | 模型管理 feature 模块 |
| api | `src/api/models.ts` | 模型 API 定义 |
| api | `src/api/newapi.ts` | NewAPI sync API 定义 |
| route page | `src/routes/models/index.tsx` | 模型页面入口 |
| route page | `src/routes/newapi/index.tsx` | NewAPI iframe 页面 |
| router | `src/router/routes.ts` | 移除 `modelsRoute`, `newapiRoute` |
| config | `src/config/routes.ts` | 移除 models、newapi 的 nav 定义 |
| app-apis | `src/api/app-apis.ts` | 移除 `modelsApi`, `newapiApi` |
| enums | `src/config/enums.ts` | 移除 `MODEL_STATUS`, `MODEL_TYPES`（如无其他引用） |

### 基础设施

| 类别 | 路径/位置 | 说明 |
|------|-----------|------|
| docker-compose | `deploy/docker-compose.prod.yml` | 移除 `newapi-sms` 服务 |
| scripts | `scripts/lib/db-reset.sh` | 移除 `sms_newapi`, `sms_logs` 数据库的创建/重置 |
| newapi 目录 | `sms/newapi/` | 整个目录（`.env.example` + bootstrap 脚本） |
| env | `sms/backend/.env.development` | 移除 `NEWAPI_*` 变量 |

### schema.sql

从 `sms/backend/schema.sql` 移除：
- `models` 表（整张）
- `sync_versions` 表
- `channels` 表（如存在，sync_pull 在用）
- `bump_sync_version_models()` 函数和触发器
- `oauth_clients` 表

### 文档

| 路径 | 动作 |
|------|------|
| `sms/docs/plan/newapi-integration.md` | 删除 |
| `sms/docs/ai-model-supplier-management-design.md` | 更新：移除模型相关章节（§4.6 models 表、API 中 /models 和 /newapi 路由、前端模型目录页面描述） |

## 保留的内容

| 模块 | 原因 |
|------|------|
| `domain/supplier/` | 供应商管理是 SMS 核心职责 |
| `domain/contract/` | 合同管理 |
| `domain/order/` | 采购订单 |
| `domain/evaluation/` | 绩效评估 |
| `domain/dashboard/` | 仪表盘（需调整：移除模型计数） |
| `domain/auth/`, `domain/user/` | SMS 自身用户体系 |
| suppliers 表的 model 关联字段 | 需确认：如果 supplier_detail 页面之前展示"该供应商的模型列表"tab，移除该 tab |

## 执行顺序

> 无需 migration、无需向后兼容（无线上数据）

1. **后端 domain 层清理** — 删除 `model/`, `newapisync/`, `sync/` 目录
2. **后端 store 层清理** — 删除 `model.go`, `sync_pull.go`, `sync_store.go`, `oauth_client_store.go`
3. **后端 integration 层清理** — 删除 `integration/newapi/`
4. **后端 handler + 路由清理** — 删除对应 handler 包，更新 `register.go`
5. **后端 deps/config/app 清理** — 移除字段和初始化逻辑
6. **schema.sql 清理** — 移除相关表定义
7. **前端清理** — 删除 features/models, api/models.ts, api/newapi.ts, routes 页面，更新 router/config/app-apis
8. **基础设施清理** — docker-compose, db-reset 脚本, `sms/newapi/` 目录
9. **文档更新** — 删除/更新 docs
10. **编译验证** — `go build ./...` + `pnpm build:sms`

## 注意事项

- `sms/backend/schema.sql` 中 `models` 表有 `supplier_id FK → suppliers`，删除 models 表后 suppliers 表不受影响
- Dashboard service 可能有 "模型总数" 的 SQL 聚合查询，需一并清理
- 供应商详情页前端如果有 "模型" tab，需移除
- `sms/frontend/src/features/query/` 中的 `queryKeys.models` 需清理
- 确认 `sms/backend/tests/` 中是否有模型/newapi 相关测试文件，一并删除

## 影响范围

- SMS 前后端仅内部使用，无外部消费者
- apps local 模式已改为从 apps SaaS Catalog API 同步（`catalogsync worker`），不再依赖 SMS sync endpoint
- `sms_newapi` 和 `sms_logs` 数据库可在清理后直接 drop

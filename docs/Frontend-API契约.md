# Frontend API 契约

本文档描述 TokenJoy 前端 `src/api/` 层调用的 REST 接口契约，供 MSW mock 与真实后端实现对齐。类型定义以 [`apps/frontend/src/api/types.ts`](../apps/frontend/src/api/types.ts) 为准。

## 通用约定

### Base URL

- 路径前缀：`{BASE_URL}/api`（本地开发通常为 `/api`）
- 常量：[`API_BASE_PATH`](../apps/frontend/src/config/app.ts)

### 请求头

| Header             | 说明                                                                          |
| ------------------ | ----------------------------------------------------------------------------- |
| `Content-Type`     | `application/json`（默认）                                                    |
| `X-Demo-Member-Id` | Demo 模式下当前成员 ID（由 `client.ts` 注入）；真实后端可替换为 session/token |

### 错误响应

HTTP 非 2xx 时，body 应包含：

```json
{ "message": "错误描述" }
```

前端 `ApiError` 会读取 `message` 字段。

### 分页

查询参数：`page`、`pageSize`（及领域特定筛选字段）

响应体 `Paginated<T>`：

```ts
{
  items: T[]
  total: number
  page: number
  pageSize: number
}
```

---

## Session

| 方法 | 路径                     | 请求              | 响应             |
| ---- | ------------------------ | ----------------- | ---------------- |
| GET  | `/session?memberId={id}` | query: `memberId` | `SessionContext` |

`SessionContext`：`{ member, permissions, readOnly }`

---

## Org（组织管理）

### 数据源 `dataSourceApi`

| 方法 | 路径                               | 请求体              | 响应                              |
| ---- | ---------------------------------- | ------------------- | --------------------------------- |
| GET  | `/org/data-source/status`          | —                   | `DataSourceStatus`                |
| POST | `/org/data-source/test`            | `Credential`        | `{ success, message? }`           |
| PUT  | `/org/data-source`                 | `Credential`        | `void`                            |
| GET  | `/org/data-source/search?keyword=` | query: `keyword`    | `{ name, department, mappingOk }` |
| POST | `/org/data-source/import`          | —                   | `ImportResult`                    |
| POST | `/org/data-source/import/retry`    | `{ ids: string[] }` | `ImportResult`                    |

### 同步 `syncApi`

| 方法 | 路径                             | 请求体                    | 响应                 |
| ---- | -------------------------------- | ------------------------- | -------------------- |
| GET  | `/org/sync/config`               | —                         | `SyncConfig`         |
| PUT  | `/org/sync/config`               | `SyncConfig`              | `void`               |
| POST | `/org/sync/trigger`              | —                         | `ImportResult`       |
| GET  | `/org/sync/logs?page=&pageSize=` | query: `page`, `pageSize` | `Paginated<SyncLog>` |

### 部门 `departmentApi`

| 方法   | 路径                    | 请求体               | 响应           |
| ------ | ----------------------- | -------------------- | -------------- |
| GET    | `/org/departments/tree` | —                    | `Department[]` |
| POST   | `/org/departments`      | `{ name, parentId }` | `Department`   |
| PUT    | `/org/departments/:id`  | `{ name }`           | `Department`   |
| DELETE | `/org/departments/:id`  | —                    | `void`         |

### 成员 `memberApi`

| 方法   | 路径                        | 请求体 / 查询                                                         | 响应                      |
| ------ | --------------------------- | --------------------------------------------------------------------- | ------------------------- |
| GET    | `/org/members`              | query: `departmentId?`, `directOnly?`, `page`, `pageSize`, `keyword?` | `Paginated<Member>`       |
| POST   | `/org/members`              | `Omit<Member, 'id' \| 'status' \| 'roles' \| 'source'>`               | `Member`                  |
| PUT    | `/org/members/:id`          | `Partial<Member>`                                                     | `Member`                  |
| DELETE | `/org/members`              | `{ ids: string[] }`                                                   | `void`                    |
| PUT    | `/org/members/status`       | `{ ids, status: 'active' \| 'inactive' }`                             | `void`                    |
| POST   | `/org/members/transfer`     | `{ ids, departmentId }`                                               | `void`                    |
| POST   | `/org/members/invite`       | `{ email?, phone? }`                                                  | `void`                    |
| POST   | `/org/members/batch-invite` | `{ ids? }`                                                            | `{ sent: number }`        |
| POST   | `/org/members/batch-import` | `{ rows: BatchImportRow[] }`                                          | `MemberBatchImportResult` |

### 角色 `roleApi`

| 方法   | 路径                                   | 请求体                  | 响应           |
| ------ | -------------------------------------- | ----------------------- | -------------- |
| GET    | `/org/roles`                           | —                       | `Role[]`       |
| POST   | `/org/roles`                           | `{ name, permissions }` | `Role`         |
| PUT    | `/org/roles/:id`                       | `{ name, permissions }` | `Role`         |
| DELETE | `/org/roles/:id`                       | —                       | `void`         |
| GET    | `/org/roles/:roleId/members`           | —                       | `Member[]`     |
| POST   | `/org/roles/:roleId/members`           | `{ memberId }`          | `void`         |
| DELETE | `/org/roles/:roleId/members/:memberId` | —                       | `void`         |
| GET    | `/org/permissions`                     | —                       | `Permission[]` |

---

## Budget（预算管理）

| 方法   | 路径                     | 请求体 / 查询                           | 响应                  |
| ------ | ------------------------ | --------------------------------------- | --------------------- |
| GET    | `/budget/tree?period=`   | query: `period?`                        | `BudgetNode[]`        |
| PUT    | `/budget/nodes/:id`      | `{ budget, reservedPool? }`             | `BudgetNode`          |
| GET    | `/budget/groups`         | —                                       | `BudgetGroup[]`       |
| POST   | `/budget/groups`         | `Omit<BudgetGroup, 'id' \| 'consumed'>` | `BudgetGroup`         |
| PUT    | `/budget/groups/:id`     | `Partial<...>`                          | `BudgetGroup`         |
| DELETE | `/budget/groups/:id`     | —                                       | `void`                |
| GET    | `/budget/overrun-policy` | —                                       | `OverrunPolicyConfig` |
| PUT    | `/budget/overrun-policy` | `OverrunPolicyConfig`                   | `OverrunPolicyConfig` |
| GET    | `/budget/alerts`         | —                                       | `AlertRule[]`         |
| POST   | `/budget/alerts`         | `Omit<AlertRule, 'id'>`                 | `AlertRule`           |
| PUT    | `/budget/alerts/:id`     | `Partial<AlertRule>`                    | `AlertRule`           |
| DELETE | `/budget/alerts/:id`     | —                                       | `void`                |

---

## Keys（API Key 管理）

### 供应商密钥 `providerKeyApi`

| 方法   | 路径                        | 请求体                    | 响应            |
| ------ | --------------------------- | ------------------------- | --------------- |
| GET    | `/keys/provider`            | —                         | `ProviderKey[]` |
| POST   | `/keys/provider`            | `{ provider, name, key }` | `ProviderKey`   |
| PUT    | `/keys/provider/:id/toggle` | `{ enabled }`             | `void`          |
| POST   | `/keys/provider/:id/rotate` | `{ newKey }`              | `ProviderKey`   |
| DELETE | `/keys/provider/:id`        | —                         | `void`          |

### 平台密钥 `platformKeyApi`

| 方法   | 路径                                     | 请求体 / 查询                                          | 响应                     |
| ------ | ---------------------------------------- | ------------------------------------------------------ | ------------------------ |
| GET    | `/keys/platform`                         | query: `page?`, `pageSize?`, `memberId?`               | `Paginated<PlatformKey>` |
| POST   | `/keys/platform`                         | `{ name, memberId?, appName?, quota, modelWhitelist }` | `PlatformKey`            |
| PUT    | `/keys/platform/:id`                     | `{ name?, quota?, modelWhitelist? }`                   | `PlatformKey`            |
| PUT    | `/keys/platform/:id/toggle`              | `{ enabled }`                                          | `PlatformKey`            |
| POST   | `/keys/platform/:id/rotate`              | —                                                      | `PlatformKey`            |
| PUT    | `/keys/platform/:id/revoke`              | —                                                      | `void`                   |
| DELETE | `/keys/platform/:id`                     | —                                                      | `void`                   |
| GET    | `/keys/platform/quota-summary?memberId=` | query: `memberId`                                      | `MemberQuotaSummary`     |

### 审批 `approvalApi`

| 方法 | 路径                              | 请求体 / 查询                                                 | 响应                                      |
| ---- | --------------------------------- | ------------------------------------------------------------- | ----------------------------------------- |
| GET  | `/keys/approvals`                 | query: `tab?`, `memberId?`                                    | `KeyApproval[]`                           |
| POST | `/keys/approvals`                 | `{ type, reason, requestedQuota, requestedModels, memberId }` | `KeyApproval`                             |
| PUT  | `/keys/approvals/:id/approve`     | —                                                             | `void`                                    |
| PUT  | `/keys/approvals/:id/reject`      | `{ reason? }`                                                 | `void`                                    |
| GET  | `/keys/approvals/:id/quota-check` | —                                                             | `{ sufficient, reservedPool, requested }` |

---

## Models（模型与路由）

### 模型 `modelApi`

| 方法 | 路径                 | 请求体             | 响应          |
| ---- | -------------------- | ------------------ | ------------- |
| GET  | `/models`            | —                  | `ModelInfo[]` |
| POST | `/models`            | `CreateModelInput` | `ModelInfo`   |
| PUT  | `/models/:id/toggle` | `{ enabled }`      | `void`        |

### 路由 `routingApi`

| 方法 | 路径                              | 请求体 / 查询                                                 | 响应                |
| ---- | --------------------------------- | ------------------------------------------------------------- | ------------------- |
| GET  | `/models/routing`                 | —                                                             | `RoutingRule[]`     |
| PUT  | `/models/routing/:id`             | `{ allowedModels, inherited, defaultModel?, fallbackModel? }` | `RoutingRule`       |
| GET  | `/models/routing/resolve?deptId=` | query: `deptId`                                               | `ResolvedWhitelist` |

---

## Dashboard（数据看板）

| 方法 | 路径                          | 查询      | 响应               |
| ---- | ----------------------------- | --------- | ------------------ |
| GET  | `/dashboard/cost/summary`     | `period?` | `CostSummary`      |
| GET  | `/dashboard/cost/departments` | —         | `DepartmentCost[]` |
| GET  | `/dashboard/cost/daily`       | `days?`   | `DailyCost[]`      |
| GET  | `/dashboard/cost/top`         | `limit?`  | `TopConsumer[]`    |
| GET  | `/dashboard/usage/models`     | —         | `ModelUsage[]`     |
| GET  | `/dashboard/usage/teams`      | —         | `TeamUsage[]`      |

---

## Audit（审计日志）

| 方法 | 路径                | 查询                                      | 响应                      |
| ---- | ------------------- | ----------------------------------------- | ------------------------- |
| GET  | `/audit/operations` | `page?`, `pageSize?`, `action?`           | `Paginated<OperationLog>` |
| GET  | `/audit/calls`      | `page?`, `pageSize?`, `model?`, `status?` | `Paginated<CallLog>`      |

---

## Mock 与真实 API 切换

### 架构

```
UI → src/api/* → client.request() → fetch(/api/...)
                                      ├─ USE_MOCKS=true  → MSW handlers（Demo API）
                                      └─ USE_MOCKS=false → 真实后端
```

Demo 模式下 `/api/*` 即 **Demo API**：由 MSW 在浏览器内实现，返回内存 fake 数据；前端 `api/*.ts` 调用方式与生产相同，不访问外部服务。

- Mock 开关：[`USE_MOCKS`](../apps/frontend/src/config/app.ts)（`DEV` 或 `VITE_ENABLE_MOCKS=true`）
- MSW handlers：[`apps/frontend/src/mocks/handlers/`](../apps/frontend/src/mocks/handlers/)
- Fixtures：[`apps/frontend/src/mocks/fixtures/`](../apps/frontend/src/mocks/fixtures/)

### 迁移到真实 API 步骤

1. **实现后端**：按本文档路径与 `types.ts` 实现 REST API
2. **关闭 Mock**：生产/预发设置 `VITE_ENABLE_MOCKS=false`（或删除该变量）
   - [`.env.production`](../apps/frontend/.env.production)
   - [`vercel.json`](../vercel.json)
   - [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml)
3. **配置网络**：
   - 静态托管（Vercel / GitHub Pages）：将 `/api` 反向代理到后端，或前后端同域部署
   - 本地开发：设置 `VITE_API_PROXY_TARGET=http://localhost:8080` 启用 Vite proxy（同时设 `VITE_ENABLE_MOCKS=false` 可全量走真实后端；保持 mock 时未匹配的请求会 bypass 到 proxy）
4. **鉴权**：在 [`client.ts`](../apps/frontend/src/api/client.ts) 将 `setDemoMemberIdProvider` 替换为真实 session/token 注入
5. **逐域验收**：session → org → budget → keys → models → dashboard → audit

### 环境变量

| 变量                    | 说明                                                                 |
| ----------------------- | -------------------------------------------------------------------- |
| `VITE_ENABLE_MOCKS`     | `true` 时生产构建也启用 MSW（当前 Demo 默认）                        |
| `VITE_API_PROXY_TARGET` | 本地开发时 Vite 将 `/api` 代理到该地址（如 `http://localhost:8080`） |

### 注意事项

- MSW mock 数据存于内存，刷新页面会重置；真实后端需持久化
- `api/*.ts` 与 `mocks/handlers/*.ts` 需保持路径同步；以本文档与 `types.ts` 为契约源
- `X-Demo-Member-Id` 在部分 handler 中尚未消费；真实后端需明确鉴权方案

# Frontend API 契约

本文档描述 TokenJoy 前端 `src/api/` 层调用的 REST 接口契约，供前后端实现对齐。

**当前状态：** Monorepo 已包含 Go 后端（[`apps/backend/`](../apps/backend/)），实现契约 §5 全部 **81** 个端点。本地联调：根目录 `pnpm start`（并发 backend + frontend，`.env.development` 代理到 `:8080`）。Dev 环境在 `/login` 选择成员并写入 `tokenjoy_session_member` cookie。前后端行为均以本文档及 `api/types/` 为准；后端设计见 [Backend-设计.md](./Backend-设计.md)。

**权威来源（按优先级）：**

1. 类型定义 — [`apps/frontend/src/api/types/`](../apps/frontend/src/api/types/)
2. HTTP 客户端 — [`apps/frontend/src/api/{domain}.ts`](../apps/frontend/src/api/)
3. Session 校验 — [`apps/frontend/src/api/schemas/session.ts`](../apps/frontend/src/api/schemas/session.ts)
4. 后端类型 — [`apps/backend/internal/domain/types/`](../apps/backend/internal/domain/types/)

文档索引见 [README.md](./README.md)；前端分层见 [Frontend-开发指南.md](./Frontend-开发指南.md)。

---

## 1. 调用架构

```
页面 / Hook
  └─ useApis() 或 useInjectedQuery({ queryFn: (apis) => ... })
       └─ AppApis.{domain}Api.{method}()
            └─ client.request() → fetch({API_BASE_PATH}{path})
                 └─ Vite dev proxy / 生产反向代理 → Go backend
```

| 模块                    | 路径                                   | 职责                                                                |
| ----------------------- | -------------------------------------- | ------------------------------------------------------------------- |
| `client.ts`             | `api/client.ts`                        | `request()`、`ApiError`、`buildQuery()`、`setUnauthorizedHandler()` |
| `app-apis.ts`           | `api/app-apis.ts`                      | `AppApis` 接口与 `defaultApis` 聚合（14 个命名空间）                |
| `api-context.ts`        | `api/api-context.ts`                   | `ApiContext` React Context                                          |
| `context.tsx`           | `api/context.tsx`                      | `ApiProvider` 注入                                                  |
| `use-apis.ts`           | `api/use-apis.ts`                      | `useApis()`、`useInjectedApis()`                                    |
| `use-injected-query.ts` | `features/query/use-injected-query.ts` | 基于 TanStack Query 的 `useInjectedQuery()`                         |
| `query-keys.ts`         | `features/query/query-keys.ts`         | 各域 query key 工厂                                                 |
| `{domain}.ts`           | `api/{domain}.ts`                      | 各资源 HTTP 方法                                                    |

**依赖注入：** 生产环境 `AdminLayout` 注入 `defaultApis`；页面 Hook 支持 `injectedApis` 参数供测试覆盖；测试通过 `createMockApis()` 注入 mock API。

**数据获取：** 读操作逐步迁移至 `useInjectedQuery` + `queryKeys`；写操作仍在事件处理函数中直接调用 `*Api` 方法，成功后通过 `queryClient.invalidateQueries` 或 `useWorkflowRefresh` 刷新缓存。

---

## 2. 通用约定

### 2.1 Base URL

| 常量            | 值               | 定义                                                  |
| --------------- | ---------------- | ----------------------------------------------------- |
| `API_BASE_PATH` | `{BASE_URL}/api` | [`config/app.ts`](../apps/frontend/src/config/app.ts) |

本地开发与 nginx 同域部署均为 `/api`。

开发环境通过 Vite dev/preview 将 `{BASE_URL}/api` 同域反代到 Go（默认 `http://127.0.0.1:8080`，可用 `VITE_API_PROXY_TARGET` 覆盖）。生产使用 nginx 等同域反代，见 [`deploy/nginx.conf.example`](../deploy/nginx.conf.example)。

### 2.2 请求头

| Header          | 必填       | 说明                                                                                                                                   |
| --------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `Accept`        | 是         | `application/json`（`client.request` 默认注入）                                                                                        |
| `Content-Type`  | 有 body 时 | `application/json`（`client.request` 默认注入）                                                                                        |
| `Cookie`        | Session    | `credentials: 'include'` 自动携带；企业面 `tokenjoy_session_member`；平台面 `tokenjoy_platform_session`（`SUPPORT_SAAS=true`，见 §10） |
| `Authorization` | 可选       | `Bearer {token}`（生产网关）                                                                                                           |

### 2.3 成功响应

- HTTP 2xx
- 返回 JSON，结构与 `api/types/` 一致
- `void` 类型端点：TypeScript 侧为 `void`；**注意** `request()` 始终调用 `res.json()`，后端对无 body 端点应返回 `{}` 或 `null`，避免空 body 导致解析失败

### 2.4 错误响应

HTTP 非 2xx 时，body 应包含：

```json
{ "message": "错误描述" }
```

前端 `ApiError`（[`client.ts`](../apps/frontend/src/api/client.ts)）读取 `message` 字段，无则回退 `statusText`。收到 **401** 时触发 `setUnauthorizedHandler` 注册的回调（`AuthUnauthorizedBridge` 跳转 `/login`）。

**常见状态码：**

| 状态码 | 场景                                                    |
| ------ | ------------------------------------------------------- |
| `400`  | 缺少必填参数、不可删预设角色等                          |
| `401`  | Session 未鉴权（`GET /session` 无有效 cookie / Bearer） |
| `404`  | 资源不存在                                              |
| `422`  | 业务校验失败                                            |
| `503`  | minute 粒度 NewAPI 不可用（可能含 `retryAfter`）        |

### 2.5 分页

**查询参数：** `page`、`pageSize`（及领域特定筛选字段）

**响应体 `Paginated<T>`：**

```ts
{
  items: T[]
  total: number
  page: number
  pageSize: number
}
```

### 2.6 查询参数构建

`buildQuery()` 会跳过 `undefined`、`null`、空字符串。布尔值序列化为 `"true"` / `"false"`。

---

## 3. 鉴权与权限

### 3.1 Session 端点

| 调用方式                  | 说明                                                                           |
| ------------------------- | ------------------------------------------------------------------------------ |
| `sessionApi.getCurrent()` | `GET /session`；凭 cookie `tokenjoy_session_member` 或 `Authorization: Bearer` |

响应经 `SessionContextSchema`（Zod）校验。成员不存在 → 404；未鉴权 → 401。加载失败由 `SessionGate` 展示错误页。

本地 dev：在 `/login` 选择成员写入 cookie 后调用 `getCurrent()`。

### 3.2 后端 Profile（`APP_PROFILE`）

| Profile | 值     | GET 读接口（demo）    | Session 身份解析                           |
| ------- | ------ | --------------------- | ------------------------------------------ |
| Demo    | `demo` | 多数 GET 免 Session   | cookie `tokenjoy_session_member` 或 Bearer |
| 生产    | `prod` | 要求 Session + 读权限 | 同上                                       |

写操作（POST / PUT / DELETE）在两种 Profile 下均要求 Session + 写权限。本地默认 `demo`；部署生产应设置 `APP_PROFILE=prod`。

### 3.3 其他端点

真实后端按 permission key 校验：未登录 → `401`；已登录但无权限 → `403`。前端通过 `GET /session` 返回的 `permissions[]` 驱动 `PermissionGate` / `usePermissions()`。

### 3.5 SaaS 双 Session

| 面     | Cookie                      | 登录入口                    | 说明                                                             |
| ------ | --------------------------- | --------------------------- | ---------------------------------------------------------------- |
| 企业面 | `tokenjoy_session_member`   | `/login` 或邀请激活         | `GET /session`；`companyId` 来自 Session，**不可**用 Header 覆盖 |
| 平台面 | `tokenjoy_platform_session` | `POST /platform/auth/login` | 与企业 Session 互不影响；`SUPPORT_SAAS=false` 时平台路由 404     |

平台控制台与企业控制台应使用独立 `ApiProvider` 或独立 `client` 配置，避免 Cookie 混用。

### 3.4 权限 Key

定义于 [`lib/permission-keys.ts`](../apps/frontend/src/lib/permission-keys.ts)：

| Key                | 含义                                            |
| ------------------ | ----------------------------------------------- |
| `org:read`         | 组织域只读（部门树、成员列表、角色等 GET）      |
| `org:datasource`   | 数据源配置与导入                                |
| `org:structure`    | 部门结构                                        |
| `org:roles`        | 角色管理                                        |
| `org:members`      | 成员管理                                        |
| `budget:read`      | 预算只读（预算树、配额、告警等 GET）            |
| `budget:allocate`  | 预算分配                                        |
| `budget:approve`   | 预算审批                                        |
| `budget:policy`    | 超支策略与告警                                  |
| `model:read`       | 模型域只读（模型列表、路由 GET）                |
| `model:manage`     | 模型管理                                        |
| `model:whitelist`  | 模型白名单 / 路由                               |
| `keys:read`        | Keys 域只读（供应商、平台 Key、审批列表等 GET） |
| `keys:admin`       | 平台 Key 管理                                   |
| `keys:provider`    | 供应商 Key 管理                                 |
| `self:keys`        | 我的 Key                                        |
| `self:approval`    | 我的审批                                        |
| `dashboard:cost`   | 成本看板                                        |
| `dashboard:usage`  | 用量看板                                        |
| `audit:read`       | 审计日志                                        |
| `api:call`         | API 调用权限                                    |
| `billing:read`     | 公司钱包只读（`GET /billing/wallet`）           |
| `billing:recharge` | 企业自助充值（`POST /billing/recharge`）        |
| `platform:*`       | 平台运营（独立 Session，不走成员权限表）        |

`readOnly: true` 表示当前 Session 无任何写权限 capability（见 `lib/permissions.ts` 中 `isReadOnlySession`）。

**生产 Profile GET 读权限映射（后端）：**

| API 前缀           | 所需读权限                                    |
| ------------------ | --------------------------------------------- |
| `/api/org/*`       | `org:read`                                    |
| `/api/budget/*`    | `budget:read`                                 |
| `/api/keys/*`      | `keys:read`                                   |
| `/api/models/*`    | `model:read`                                  |
| `/api/audit/*`     | `audit:read`                                  |
| `/api/dashboard/*` | `dashboard:cost` 与 `dashboard:usage`（均需） |

审计员（`RoleAuditor`）预设角色包含上述读权限，可用于只读浏览。

---

## 4. 共享类型

定义于 [`api/types/common.ts`](../apps/frontend/src/api/types/common.ts)。

### Paginated\<T\>

见 §2.5。

### SessionContext

| 字段          | 类型       | 说明                                |
| ------------- | ---------- | ----------------------------------- |
| `member`      | `Member`   | 当前成员                            |
| `permissions` | `string[]` | 权限 key 列表                       |
| `readOnly`    | `boolean`  | 无写权限时为 `true`                 |
| `companyId`   | `number`   | 当前成员所属企业 ID；私有化默认 `1` |

---

## 5. 端点清单

路径均相对于 `API_BASE_PATH`（`/api`）。共 **81** 个端点。类型详情见 §6。

### 5.1 Session

客户端：[`sessionApi`](../apps/frontend/src/api/session.ts)

| 方法 | 路径       | 查询 / Body | 响应             | 说明                                   |
| ---- | ---------- | ----------- | ---------------- | -------------------------------------- |
| GET  | `/session` | —           | `SessionContext` | Cookie / Bearer 解析成员；未鉴权 → 401 |

---

### 5.2 Org（组织管理）

客户端：[`org.ts`](../apps/frontend/src/api/org.ts)

#### 数据源 `dataSourceApi`

| 方法 | 路径                            | Body / 查询         | 响应                                     |
| ---- | ------------------------------- | ------------------- | ---------------------------------------- |
| GET  | `/org/data-source/status`       | —                   | `DataSourceStatus`                       |
| POST | `/org/data-source/test`         | `Credential`        | `{ success: boolean; message?: string }` |
| PUT  | `/org/data-source`              | `Credential`        | `void`                                   |
| GET  | `/org/data-source/search`       | query: `keyword`    | `{ name, department, mappingOk }`        |
| POST | `/org/data-source/import`       | —                   | `ImportResult`                           |
| POST | `/org/data-source/import/retry` | `{ ids: string[] }` | `ImportResult`                           |

#### 同步 `syncApi`

| 方法 | 路径                | Body / 查询               | 响应                 |
| ---- | ------------------- | ------------------------- | -------------------- |
| GET  | `/org/sync/config`  | —                         | `SyncConfig`         |
| PUT  | `/org/sync/config`  | `SyncConfig`              | `void`               |
| POST | `/org/sync/trigger` | —                         | `ImportResult`       |
| GET  | `/org/sync/logs`    | query: `page`, `pageSize` | `Paginated<SyncLog>` |

#### 部门 `departmentApi`

| 方法   | 路径                    | Body                 | 响应           |
| ------ | ----------------------- | -------------------- | -------------- |
| GET    | `/org/departments/tree` | —                    | `Department[]` |
| POST   | `/org/departments`      | `{ name, parentId }` | `Department`   |
| PUT    | `/org/departments/:id`  | `{ name }`           | `Department`   |
| DELETE | `/org/departments/:id`  | —                    | `void`         |

#### 成员 `memberApi`

| 方法   | 路径                        | Body / 查询                                                            | 响应                      |
| ------ | --------------------------- | ---------------------------------------------------------------------- | ------------------------- |
| GET    | `/org/members`              | query: `departmentId?`, `directOnly?`, `page`, `pageSize`, `keyword?`  | `Paginated<Member>`       |
| POST   | `/org/members`              | `Omit<Member, 'id' \| 'status' \| 'roles' \| 'source' \| 'companyId'>` | `Member`                  |
| PUT    | `/org/members/:id`          | `Partial<Member>`                                                      | `Member`                  |
| DELETE | `/org/members`              | `{ ids: string[] }`                                                    | `void`                    |
| PUT    | `/org/members/status`       | `{ ids, status: 'active' \| 'inactive' }`                              | `void`                    |
| POST   | `/org/members/transfer`     | `{ ids, departmentId }`                                                | `void`                    |
| POST   | `/org/members/invite`       | `{ email?, phone? }`                                                   | `void`                    |
| POST   | `/org/members/batch-invite` | `{ ids? }`                                                             | `{ sent: number }`        |
| POST   | `/org/members/batch-import` | `{ rows: BatchImportRow[] }`                                           | `MemberBatchImportResult` |

#### 角色 `roleApi`

| 方法   | 路径                                   | Body                              | 响应           |
| ------ | -------------------------------------- | --------------------------------- | -------------- |
| GET    | `/org/roles`                           | —                                 | `Role[]`       |
| POST   | `/org/roles`                           | `{ name, permissions: string[] }` | `Role`         |
| PUT    | `/org/roles/:id`                       | `{ name, permissions: string[] }` | `Role`         |
| DELETE | `/org/roles/:id`                       | —                                 | `void`         |
| GET    | `/org/roles/:roleId/members`           | —                                 | `Member[]`     |
| POST   | `/org/roles/:roleId/members`           | `{ memberId }`                    | `void`         |
| DELETE | `/org/roles/:roleId/members/:memberId` | —                                 | `void`         |
| GET    | `/org/permissions`                     | —                                 | `Permission[]` |

---

### 5.3 Budget（预算管理）

客户端：[`budgetApi`](../apps/frontend/src/api/budget.ts)

| 方法   | 路径                                        | Body / 查询                                      | 响应                  | 备注                              |
| ------ | ------------------------------------------- | ------------------------------------------------ | --------------------- | --------------------------------- |
| GET    | `/budget/tree`                              | query: `period?`                                 | `BudgetNode[]`        |                                   |
| PUT    | `/budget/nodes/:id`                         | `{ budget, reservedPool? }`                      | `BudgetNode`          |                                   |
| GET    | `/budget/departments/:deptId/member-quotas` | —                                                | `MemberBudgetQuota[]` |                                   |
| PUT    | `/budget/members/:memberId`                 | `UpdateMemberQuotaInput`                         | `MemberBudgetQuota`   | 部门内超卖 / 低于已分配 Key → 422 |
| GET    | `/budget/groups`                            | —                                                | `BudgetGroup[]`       |                                   |
| POST   | `/budget/groups`                            | `Omit<BudgetGroup, 'id' \| 'consumed'>`          | `BudgetGroup`         |                                   |
| PUT    | `/budget/groups/:id`                        | `Partial<Omit<BudgetGroup, 'id' \| 'consumed'>>` | `BudgetGroup`         |                                   |
| DELETE | `/budget/groups/:id`                        | —                                                | `void`                |                                   |
| GET    | `/budget/overrun-policy`                    | —                                                | `OverrunPolicyConfig` |                                   |
| PUT    | `/budget/overrun-policy`                    | `OverrunPolicyConfig`                            | `OverrunPolicyConfig` |                                   |
| GET    | `/budget/alerts`                            | —                                                | `AlertRule[]`         |                                   |
| POST   | `/budget/alerts`                            | `Omit<AlertRule, 'id'>`                          | `AlertRule`           |                                   |
| PUT    | `/budget/alerts/:id`                        | `Partial<AlertRule>`                             | `AlertRule`           |                                   |
| DELETE | `/budget/alerts/:id`                        | —                                                | `void`                |                                   |

---

### 5.4 Keys（API Key 管理）

客户端：[`keys.ts`](../apps/frontend/src/api/keys.ts)

#### 供应商密钥 `providerKeyApi`

| 方法   | 路径                        | Body                      | 响应            |
| ------ | --------------------------- | ------------------------- | --------------- |
| GET    | `/keys/provider`            | —                         | `ProviderKey[]` |
| POST   | `/keys/provider`            | `{ provider, name, key }` | `ProviderKey`   |
| PUT    | `/keys/provider/:id/toggle` | `{ enabled }`             | `void`          |
| POST   | `/keys/provider/:id/rotate` | `{ newKey }`              | `ProviderKey`   |
| DELETE | `/keys/provider/:id`        | —                         | `void`          |

#### 平台密钥 `platformKeyApi`

| 方法   | 路径                           | Body / 查询                                                            | 响应                     | 备注                                                                 |
| ------ | ------------------------------ | ---------------------------------------------------------------------- | ------------------------ | -------------------------------------------------------------------- |
| GET    | `/keys/platform`               | query: `page?`, `pageSize?`, `memberId?`, `budgetGroupId?`             | `Paginated<PlatformKey>` | `memberId` / `budgetGroupId` 过滤                                    |
| POST   | `/keys/platform`               | `{ name, memberId?, budgetGroupId?, appName?, quota, modelWhitelist }` | `PlatformKey`            | 个人 Key 缺 `memberId` → 400；Group Key 校验组剩余额度；白名单 → 422 |
| PUT    | `/keys/platform/:id`           | `{ name?, quota?, modelWhitelist? }`                                   | `PlatformKey`            | 额度 / 白名单校验 → 422                                              |
| PUT    | `/keys/platform/:id/toggle`    | `{ enabled }`                                                          | `PlatformKey`            |                                                                      |
| POST   | `/keys/platform/:id/rotate`    | —                                                                      | `PlatformKey`            | 响应含 `fullKey`                                                     |
| PUT    | `/keys/platform/:id/revoke`    | —                                                                      | `void`                   |                                                                      |
| DELETE | `/keys/platform/:id`           | —                                                                      | `void`                   |                                                                      |
| GET    | `/keys/platform/quota-summary` | query: `memberId`                                                      | `MemberQuotaSummary`     |                                                                      |

#### 审批 `approvalApi`

| 方法 | 路径                              | Body / 查询                                                   | 响应                                      | 备注                                |
| ---- | --------------------------------- | ------------------------------------------------------------- | ----------------------------------------- | ----------------------------------- |
| GET  | `/keys/approvals`                 | query: `tab?`, `memberId?`                                    | `KeyApproval[]`                           | `tab`: `pending` \| `mine` \| `all` |
| POST | `/keys/approvals`                 | `{ type, reason, requestedQuota, requestedModels, memberId }` | `KeyApproval`                             | 白名单校验 → 422                    |
| PUT  | `/keys/approvals/:id/approve`     | —                                                             | `void`                                    | 预留池不足 → 422                    |
| PUT  | `/keys/approvals/:id/reject`      | `{ reason? }`                                                 | `void`                                    |                                     |
| GET  | `/keys/approvals/:id/quota-check` | —                                                             | `{ sufficient, reservedPool, requested }` |                                     |

---

### 5.5 Models（模型与路由）

客户端：[`models.ts`](../apps/frontend/src/api/models.ts)

#### 模型 `modelApi`

| 方法 | 路径                 | Body               | 响应          |
| ---- | -------------------- | ------------------ | ------------- |
| GET  | `/models`            | —                  | `ModelInfo[]` |
| POST | `/models`            | `CreateModelInput` | `ModelInfo`   |
| PUT  | `/models/:id/toggle` | `{ enabled }`      | `void`        |

#### 路由 `routingApi`

| 方法 | 路径                      | Body / 查询                                                   | 响应                |
| ---- | ------------------------- | ------------------------------------------------------------- | ------------------- |
| GET  | `/models/routing`         | —                                                             | `RoutingRule[]`     |
| PUT  | `/models/routing/:id`     | `{ allowedModels, inherited, defaultModel?, fallbackModel? }` | `RoutingRule`       |
| GET  | `/models/routing/resolve` | query: `deptId`                                               | `ResolvedWhitelist` |

---

### 5.6 Dashboard（数据看板）

客户端：[`dashboardApi`](../apps/frontend/src/api/dashboard.ts)

**只读约束：** 本节 **全部端点为 `GET`**，查询用量/成本，**无副作用**（不写库、不触发 ingest、不修改预算）。实现架构见 [Backend-设计.md](./Backend-设计.md) §10。

成本类接口共享查询参数 `CostQueryParams`（`period` 默认 `current_month`）：

| 字段          | 含义                                                                         |
| ------------- | ---------------------------------------------------------------------------- |
| `period`      | `current_month` \| `last_month` \| `last_7_days` \| `custom`                 |
| `startDate`   | `period=custom` 时必填，ISO 日期（按响应 `timezone` 解释）                   |
| `endDate`     | `period=custom` 时必填，ISO 日期                                             |
| `granularity` | 趋势粒度：`day` \| `hour` \| `week` \| `month`（`minute` 仅 `usage/series`） |

**数据源约定**

- `cost/*`、`usage/models`、`usage/teams` 的 **consumed / cost** 均来自 **`usage_buckets` 周期聚合**（不读 `budget_nodes.consumed`）。
- `usage/teams` 的 **quota** 来自 **`budget_nodes`** 预算树。
- 聚合/展示时区默认 **`Asia/Shanghai`**（IANA）；响应 `timezone` 字段返回实际使用值；企业可配置覆盖。
- `week` / `month` 由服务端对 buckets 做 `date_trunc('week' \| 'month', …)`，不走 `usage/series`。

**用量时间序列（Phase 3）** — 统一 `day` / `hour` / `minute` 查询形状：

| 方法 | 路径                      | 查询               | 响应                  |
| ---- | ------------------------- | ------------------ | --------------------- |
| GET  | `/dashboard/usage/series` | `UsageSeriesQuery` | `UsageSeriesResponse` |

`UsageSeriesQuery`：`granularity`（`day` \| `hour` \| `minute`，必填）、`start`、`end`（ISO8601 或 `YYYY-MM-DD`）、`groupBy?`（`none` \| `department` \| `member` \| `model`，**单选**）、`departmentId?`、`memberId?`

| 路径           | 数据源                     | 说明                                                                                                                                                                                             |
| -------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `day` / `hour` | `usage_buckets`            | `source: "buckets"`，`approximate: false`，`mappingAsOf: "ingest_time"`                                                                                                                          |
| `minute`       | NewAPI `ListLogs` 只读聚合 | `source: "logs"`，`approximate: true`，`mappingAsOf: "query_time"`；**最大窗口 3h**；禁止与 hour/day 混合环比；NewAPI 不可用返回 **503**，error body 可含 `retryAfter`（秒，建议客户端退避重试） |

**`groupBy` 与响应上限：** `none` 时每个 time bucket 聚合成单点；`len(points) ≤ 10000` 超限 **422**。

**cost 端点 `granularity`：** `cost/daily` 等支持 `day` \| `hour` \| `week` \| `month`（服务端 buckets `date_trunc`）；`minute` 传 cost 端点返回 **400**。

| 方法 | 路径                                          | 查询                            | 响应                     |
| ---- | --------------------------------------------- | ------------------------------- | ------------------------ |
| GET  | `/dashboard/cost/summary`                     | `CostQueryParams`               | `CostSummary`            |
| GET  | `/dashboard/cost/departments`                 | `parentId?` + `CostQueryParams` | `DepartmentCost[]`       |
| GET  | `/dashboard/cost/departments/:deptId/members` | `CostQueryParams`               | `DepartmentCostMember[]` |
| GET  | `/dashboard/cost/daily`                       | `CostQueryParams`               | `DailyCost[]`            |
| GET  | `/dashboard/cost/top`                         | `limit?` + `CostQueryParams`    | `TopConsumer[]`          |
| GET  | `/dashboard/usage/models`                     | `CostQueryParams?`              | `ModelUsage[]`           |
| GET  | `/dashboard/usage/teams`                      | `CostQueryParams?`              | `TeamUsage[]`            |

`parentId` 为空时返回顶层部门成本；指定时返回该部门的子部门成本列表，用于成本钻取。

---

### 5.7 Audit（审计日志）

客户端：[`auditApi`](../apps/frontend/src/api/audit.ts)

| 方法 | 路径                | 查询 / Body                  | 响应                      |
| ---- | ------------------- | ---------------------------- | ------------------------- |
| GET  | `/audit/settings`   | —                            | `AuditSettings`           |
| PUT  | `/audit/settings`   | `AuditSettings`              | `AuditSettings`           |
| GET  | `/audit/operations` | `AuditOperationsQueryParams` | `Paginated<OperationLog>` |
| GET  | `/audit/calls`      | `AuditCallsQueryParams`      | `Paginated<CallLog>`      |

`AuditOperationsQueryParams`：`page?`, `pageSize?`, `action?`, `from?`, `to?`, `operatorId?`, `keyword?`

`AuditCallsQueryParams`：`page?`, `pageSize?`, `model?`, `status?`, `from?`, `to?`, `callerId?`, `keyword?`

`action` 过滤值见 `AuditAction`；`status` 过滤值：`success` \| `error` \| `filtered`。

---

## 6. 类型参考

完整定义以源码为准。以下列出各域核心类型字段。

### 6.1 Org — [`types/org.ts`](../apps/frontend/src/api/types/org.ts)

**Credential**（discriminated union by `platform`）：

| platform   | 字段                            |
| ---------- | ------------------------------- |
| `feishu`   | `appId`, `appSecret`            |
| `dingtalk` | `corpId`, `appKey`, `appSecret` |
| `wecom`    | `corpId`, `secret`, `agentId`   |

**DataSourceStatus：** `platform`, `connected`, `lastImport`, `lastImportResult`

**ImportResult：** `successMembers`, `successDepartments`, `failures: ImportFailure[]`

**ImportFailure：** `id`, `name`, `employeeId`, `reason`

**SyncConfig：** `enabled`, `startTime`, `frequencyHours`（`6` \| `12` \| `24`）, `deleteMemberThreshold`, `deleteDepartmentThreshold`, `notifyPhone`, `notifyEmail`, `notifyIm`

**SyncLog：** `id`, `time`, `type`（`scheduled` \| `manual`）, `result`（`success` \| `partial_failure` \| `failure`）, `detail`

**Department：** `id`, `name`, `parentId`, `children?`, `memberCount`

**Member：** `id`, `companyId`, `name`, `phone`, `email`, `departmentId`, `departmentName`, `status`（`active` \| `inactive` \| `pending`）, `roles`, `source`（`imported` \| `manual` \| `invited`）

**BatchImportRow：** `name`, `phone`, `email`, `departmentName`

**MemberBatchImportResult：** `imported`, `failures: { row, reason }[]`

**Role：** `id`, `name`, `type`（`preset` \| `custom`）, `permissions`, `memberCount`

**Permission：** `id`, `name`, `group`

### 6.2 Budget — [`types/budget.ts`](../apps/frontend/src/api/types/budget.ts)

**BudgetNode：** `id`, `name`, `parentId`, `budget`, `consumed`, `reservedPool?`, `children?`, `period`

**MemberBudgetQuota：** `memberId`, `memberName`, `departmentId`, `personalQuota`, `allocated`, `used`

**UpdateMemberQuotaInput：** `personalQuota`

**BudgetGroup：** `id`, `name`, `budget`, `consumed`, `memberIds`, `departmentIds`

**OverrunPolicyConfig：** `thresholds`, `notifyEmail`, `notifyPhone`, `notifyIm`, `blockMessage`

**AlertRule：** `id`, `nodeId`, `nodeName`, `thresholds`, `notifyRoleIds`, `enabled`

**CreateModelInput：** `name`, `displayName`, `baseUrl`, `apiKey`, `inputPrice`, `outputPrice`

**ResolvedWhitelist：** `inherited`, `allowedModels`, `parentCount`

### 6.3 Keys — [`types/keys.ts`](../apps/frontend/src/api/types/keys.ts)

**ProviderType：** `openai` \| `anthropic` \| `deepseek` \| `qwen` \| `custom`

**KeyStatus：** `active` \| `disabled` \| `expired` \| `error`

**ProviderKey：** `id`, `provider`, `name`, `keyPrefix`, `status`, `balance`, `lastUsed`, `createdAt`, `rotateEnabled`

**PlatformKey：** `id`, `name`, `keyPrefix`, `fullKey?`, `memberId`, `memberName`, `appName`, `budgetGroupId`, `budgetGroupName`, `status`, `quota`, `used`, `modelWhitelist`, `createdAt`, `expiresAt`

**ApprovalType：** `key` \| `quota` · **ApprovalStatus：** `pending` \| `approved` \| `rejected`

**KeyApproval：** `id`, `type`, `applicant`, `applicantId`, `department`, `reason`, `requestedQuota`, `requestedModels`, `status`, `approver`, `rejectReason?`, `createdAt`, `resolvedAt`

**MemberQuotaSummary：** `totalQuota`, `used`, `remaining`, `reservedPool`

### 6.4 Models — [`types/models.ts`](../apps/frontend/src/api/types/models.ts)

**ModelInfo：** `id`, `provider`, `name`, `displayName`, `inputPrice`, `outputPrice`, `maxContext`, `enabled`, `capabilities`

**RoutingRule：** `id`, `nodeId`, `nodeName`, `allowedModels`, `defaultModel`, `fallbackModel`, `inherited`

### 6.5 Dashboard — [`types/dashboard.ts`](../apps/frontend/src/api/types/dashboard.ts)

**CostPeriod：** `current_month` \| `last_month` \| `last_7_days` \| `custom`

**CostGranularity：** `day` \| `hour` \| `week` \| `month`

**UsageGranularity：** `day` \| `hour` \| `minute`

**UsageSeriesGroupBy：** `none` \| `department` \| `member` \| `model`

**UsageSeriesSource：** `buckets` \| `logs`

**UsageMappingAsOf：** `ingest_time` \| `query_time`

**CostQueryParams：** `period?`, `startDate?`, `endDate?`, `granularity?`

**UsageSeriesQuery：** `granularity`, `start`, `end`, `groupBy?`, `departmentId?`, `memberId?`

**UsageSeriesPoint：** `bucket`（ISO8601 时间桶起点，含时区偏移）, `departmentId?`, `memberId?`, `model?`, `costCny`, `callCount`, `inputTokens`, `outputTokens`

**UsageSeriesResponse：** `granularity`, `source`, `timezone`（默认 `Asia/Shanghai`）, `approximate`, `mappingAsOf`, `unmappedCount?`, `truncated?`, `points: UsageSeriesPoint[]`

| 字段            | buckets 路径    | minute 路径                   |
| --------------- | --------------- | ----------------------------- |
| `source`        | `"buckets"`     | `"logs"`                      |
| `approximate`   | `false`         | `true`                        |
| `mappingAsOf`   | `"ingest_time"` | `"query_time"`                |
| `unmappedCount` | 省略或 `0`      | ≥ `0`（无 mapping 的 log 数） |
| `truncated`     | 省略或 `false`  | 分页触顶时为 `true`           |

Phase 3：`inputTokens` / `outputTokens` 恒为 `0`（webhook 未带 token 字段）。

**CostSummary：** `totalCost`, `totalCostMom`, `totalTokens`, `totalRequests`, `totalRequestsMom`, `avgCostPerRequest`, `avgCostPerRequestMom`, `avgCostPerMember`, `avgCostPerMemberMom`

**DepartmentCost：** `departmentId`, `departmentName`, `cost`, `percentage`, `hasChildren?`

**DepartmentCostMember：** `memberId`, `memberName`, `cost`, `requests`, `tokens`

**DailyCost：** `date`, `cost`, `tokens`, `requests`

**TopConsumer：** `memberId`, `memberName`, `department`, `cost`, `tokens`, `requests`

**ModelUsage：** `modelId`, `modelName`, `provider`, `requests`, `tokens`, `cost`, `percentage`

**TeamUsage：** `departmentId`, `departmentName`, `quota`, `consumed`, `memberCount`, `topModel`

### 6.6 Audit — [`types/audit.ts`](../apps/frontend/src/api/types/audit.ts)

**AuditAction：** `key_create` \| `key_disable` \| `key_rotate` \| `budget_change` \| `budget_approve` \| `permission_change` \| `role_assign` \| `model_whitelist_change` \| `member_add` \| `member_remove` \| `org_structure_change`

**OperationLog：** `id`, `action`, `operator`, `operatorId`, `target`, `detail`, `ip`, `createdAt`

**CallLog：** `id`, `caller`, `callerId`, `callerType`（`member` \| `platform_key`）, `model`, `provider`, `inputTokens`, `outputTokens`, `latencyMs`, `status`, `cost`, `createdAt`, `inputPreview`, `outputPreview`

**AuditSettings：** `contentRetentionEnabled`

---

## 7. AppApis 聚合

[`app-apis.ts`](../apps/frontend/src/api/app-apis.ts) 中 `defaultApis` 包含 14 个命名空间（SaaS 落地后预计增加 `authApi`、`billingApi`、`platformApi`，见 §10）：

`sessionApi`, `dataSourceApi`, `syncApi`, `departmentApi`, `memberApi`, `roleApi`, `budgetApi`, `providerKeyApi`, `platformKeyApi`, `approvalApi`, `modelApi`, `routingApi`, `dashboardApi`, `auditApi`

所有 HTTP 调用均经 `client.request()`，无其他 `fetch('/api/...')` 直连。

---

## 8. 开发与部署

### 8.1 环境变量

| 变量                    | 说明                                                     |
| ----------------------- | -------------------------------------------------------- |
| `VITE_API_PROXY_TARGET` | 可选；覆盖 Vite 反代目标（默认 `http://127.0.0.1:8080`） |
| `BASE_URL`              | Vite 应用根路径；影响 `API_BASE_PATH`（默认 `/`）        |

根目录 `pnpm start` 并发启动 backend + frontend；[`apps/frontend/.env.development`](../apps/frontend/.env.development) 默认配置代理目标。

### 8.2 鉴权

- **企业面 Dev**：访问 [`/login`](../apps/frontend/src/routes/auth/login.tsx) 选择成员，写入 `tokenjoy_session_member` cookie
- **企业面生产**：网关或后端签发 `Authorization: Bearer` token
- **平台面（SaaS）**：`POST /platform/auth/login` → `tokenjoy_platform_session`；与企业面分离
- **邀请激活（SaaS）**：`POST /auth/accept-invite` 设密并创建成员 Session
- 401 时 `AuthUnauthorizedBridge` 跳转对应登录页（企业 `/login`；平台 `/platform/login`）

### 8.3 联调验收

1. 启动 `pnpm start` 或分别启动 backend / frontend
2. 登录后逐域抽查：session → org → budget → keys → models → dashboard → audit
3. 看板 `/dashboard/cost` 切换 `granularity`（day/week/month）应即时刷新

---

## 9. 变更检查清单

新增或修改 API 时，同步更新：

- [ ] `api/{domain}.ts` — HTTP 方法
- [ ] `api/types/{domain}.ts` — 请求 / 响应类型
- [ ] `api/app-apis.ts` — 若新增域 API 对象
- [ ] `api/schemas/` — 若响应需运行时校验（如 Session）
- [ ] `features/query/query-keys.ts` — 若读操作使用 React Query
- [ ] 本文档 §5 端点清单（SaaS 见 §10）
- [ ] 后端 handler + domain service（[`apps/backend/`](../apps/backend/)）
- [ ] 页面 Hook / 测试 — 按需补充

---

## 10. SaaS 多企业

产品模型：**企业（Company）** = 一家公司；**成员（User）** = 企业内员工。计费双轴（公司钱包 + 部门 budget）、Gateway、平台鉴权详见 [Backend-SaaS多租户改造.md](./Backend-SaaS多租户改造.md)。

**共 10 个端点**（路径均相对于 `API_BASE_PATH`）。

### 10.1 企业面：认证

客户端：`authApi`（待实现）

| 方法 | 路径                  | Body                         | 响应             | 说明                                                                    |
| ---- | --------------------- | ---------------------------- | ---------------- | ----------------------------------------------------------------------- |
| POST | `/auth/accept-invite` | `{ token, password, name? }` | `SessionContext` | 邀请激活；写入 `tokenjoy_session_member`；`token` 一次性、默认 7 天有效 |

无需 Session；成功后与 `GET /session` 结构一致（含 `companyId`）。

### 10.2 企业面：计费

客户端：`billingApi`（待实现）

| 方法 | 路径                | Body / 查询                  | 响应            | 权限               | 说明                                |
| ---- | ------------------- | ---------------------------- | --------------- | ------------------ | ----------------------------------- |
| GET  | `/billing/wallet`   | —                            | `WalletSummary` | `billing:read`     | 读 NewAPI 公司钱包                  |
| POST | `/billing/recharge` | `{ amount, idempotencyKey }` | `RechargeOrder` | `billing:recharge` | 创建 `pending` 订单；支付回调后入账 |

**`WalletSummary`**

| 字段          | 类型     | 说明                                 |
| ------------- | -------- | ------------------------------------ |
| `balance`     | `number` | 公司钱包余额（CNY）                  |
| `allocatable` | `number` | 可分配到 Token 的上限（≤ `balance`） |
| `currency`    | `'CNY'`  | 固定                                 |
| `companyId`   | `number` | 企业 ID                              |

**`RechargeOrder`**

| 字段        | 类型                                             | 说明               |
| ----------- | ------------------------------------------------ | ------------------ |
| `id`        | `string`                                         | 订单 ID            |
| `companyId` | `number`                                         | 企业               |
| `amount`    | `number`                                         | CNY                |
| `status`    | `'pending' \| 'paid' \| 'topped_up' \| 'failed'` | 见后端 §3.4 状态机 |
| `source`    | `string`                                         | `self` / 支付渠道  |
| `createdAt` | `string`                                         | ISO 8601           |

充值**不**自动涨部门 `budget`；入账后前端应引导超管进入预算页分配（见后端 §4.1.1）。

### 10.3 平台面：认证

客户端：`platformApi`

| 方法 | 路径                   | Body                  | 响应              | 说明                             |
| ---- | ---------------------- | --------------------- | ----------------- | -------------------------------- |
| POST | `/platform/auth/login` | `{ email, password }` | `PlatformSession` | 写入 `tokenjoy_platform_session` |

**`PlatformSession`**

| 字段       | 类型                            | 说明         |
| ---------- | ------------------------------- | ------------ |
| `operator` | `{ id: string; email: string }` | 平台运营账号 |

其余 `/platform/*` 须携带平台 Session；`SUPPORT_SAAS=false` 时返回 404。

### 10.4 平台面：企业与 Channel

| 方法  | 路径                               | Body / 查询                     | 响应                 | 说明                            |
| ----- | ---------------------------------- | ------------------------------- | -------------------- | ------------------------------- |
| GET   | `/platform/companies`              | `page?`, `pageSize?`, `status?` | `Paginated<Company>` | 企业列表                        |
| POST  | `/platform/companies`              | `CreateCompanyRequest`          | `Company`            | 开户 + 发超管邀请               |
| PATCH | `/platform/companies/:id`          | `PatchCompanyRequest`           | `Company`            | 状态、套餐                      |
| POST  | `/platform/companies/:id/recharge` | `{ amount }`                    | `RechargeOrder`      | 平台代充；`source=platform`     |
| GET   | `/platform/channels`               | —                               | `ProviderKey[]`      | 全局上游 Channel 列表           |
| POST  | `/platform/channels`               | 同企业面 `provider-keys` 创建体 | `ProviderKey`        | 同步到 NewAPI `platform_shared` |

**`Company`**

| 字段            | 类型                      | 说明                   |
| --------------- | ------------------------- | ---------------------- |
| `id`            | `number`                  | `companyId`            |
| `slug`          | `string`                  | 子域名标识             |
| `name`          | `string`                  | 公司名                 |
| `status`        | `'active' \| 'suspended'` | 停用后该企业 Relay 403 |
| `packageId`     | `string \| null`          | MVP 仅展示             |
| `walletBalance` | `number \| null`          | 列表可选展示           |
| `createdAt`     | `string`                  | ISO 8601               |

**`CreateCompanyRequest`：** `name`, `slug`, `superAdminEmail`, `packageId?`

**`PatchCompanyRequest`：** `status?`, `packageId?`, `name?`

> 路径以 `companies` 为准。

### 10.5 既有端点的 SaaS 行为差异

| 端点 / 域                      | 私有化          | SaaS                                                |
| ------------------------------ | --------------- | --------------------------------------------------- |
| `GET /session`                 | `companyId = 1` | `companyId` = 成员所属企业                          |
| `provider-keys` 写             | 允许            | **403**                                             |
| `POST /keys/platform`          | 不变            | 不变；后端 Token 绑企业服务账户 + `platform_shared` |
| 组织 / 预算 / 密钥读写在企业面 | 单企业          | 按 Session `companyId` 隔离                         |
| 成员邀请（企业内）             | 可选            | `POST /org/members/invite`（待定义）                |

### 10.6 前端落地检查清单

- [ ] `api/auth.ts`、`api/billing.ts`、`api/platform.ts` + 类型
- [ ] `app-apis.ts` 注册新命名空间
- [ ] 平台控制台独立路由与 `PlatformSessionGate`
- [ ] 邀请激活页 `/invite/accept?token=`
- [ ] 充值成功 → 预算分配引导（空状态文案对齐 §4.1.1）
- [ ] `permission-keys.ts` 增加 `billing:*`

后端详案：[Backend-SaaS多租户改造.md](./Backend-SaaS多租户改造.md)。NewAPI 部署：[NewAPI-SaaS多企业配置.md](./NewAPI-SaaS多企业配置.md)。

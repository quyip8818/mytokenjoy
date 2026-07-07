# TokenJoy Frontend

`apps/frontend` 现状：架构、API 契约、本地联调。后端见 [Backend.md](./Backend.md)（索引）；工程待办见 [plan.md](./plan.md)；产品差距见 [Roadmap.md](./Roadmap.md)。

**权威来源：** API 路径与 JSON → 本文 §5 + `apps/frontend/src/api/types/`。

---

## 1. 技术栈与运行时

React 19、React Router 7、Vite 8、TypeScript 5.x、Tailwind CSS 4、Radix + shadcn、TanStack Table、Recharts、react-hook-form、Zustand、TanStack Query、Vitest。CI：Node 24、Go 1.24。

路径别名：`@/` → `src/`，`@tests/` → `tests/`。

```
main.tsx → App.tsx
└─ AppProviders（ApiProvider + QueryProvider + AuthSessionProvider）
   └─ AdminLayout / MemberLayout
      ├─ AuthUnauthorizedBridge + SessionNavigationBridge
      └─ WorkflowProvider → Sidebar / Header / Outlet / WorkflowPanelStack
```

- `pnpm start`：backend `:8080` + Vite `:5173`；`/api` 反代 Go（`vite-api-proxy.ts`）
- 登录：`POST /auth/login`（邮箱密码）→ HttpOnly JWT Cookie → `GET /session`
- 首页 `/` 按权限跳转 `HOME_PATH_CANDIDATES`

---

## 2. 目录与分层

```
apps/frontend/
├── tests/                  Vitest（`tests/features/` 为主；`tests/routes/` 遗留）
└── src/
    ├── config/             routes.ts、nav.ts、app.ts
    ├── features/{domain}/  hooks/、components/、lib/（canonical 页面逻辑）
    ├── routes/{domain}/    薄页面入口（从 @/features/* 导入）
    ├── components/         ui/、layout/、auth/、{domain}/（遗留，待迁入 features）
    ├── api/                client + 域 API + types/
    ├── hooks/
    └── lib/
```

| 代码       | 位置（目标态）                                               |
| ---------- | ------------------------------------------------------------ |
| 页面入口   | `routes/{domain}/{page}.tsx`                                 |
| 页面逻辑   | `features/{domain}/hooks/use-{page}-page.ts`                 |
| 页面 Shell | `features/{domain}/components/*-page-shell.tsx`              |
| 域内 UI    | `features/{domain}/components/`                              |
| 遗留 UI    | `components/{domain}/`（迁移中，见 [plan.md](./plan.md) §4） |
| HTTP / DTO | `api/{domain}.ts`、`api/types/`                              |
| 纯逻辑     | `lib/`、`features/{domain}/lib/`                             |

禁止硬编码路由（用 `ROUTES.*`）；`components/ui` 不含业务语义。工程待办见 [plan.md](./plan.md) §4。

### 2.1 约定与门禁

`scripts/check-conventions.ts` 在 CI/lint 中强制执行：

- 页面逻辑在 `features/{domain}/hooks/use-*-page.ts`；**禁止**在 `components/{domain}/` 里直接 `useApis()`
- `components/ui` 不得出现领域名（budget、org 等）
- `components/` 不得反向依赖 `@/routes/`
- 禁止 `../../` 及更深的相对路径，统一 `@/` 别名
- 路由 `lazy` 目标文件必须存在

**组件归属：**

| 场景           | 放置位置                                                                   |
| -------------- | -------------------------------------------------------------------------- |
| 多路由复用     | `features/{domain}/components/`（目标）；遗留可能在 `components/{domain}/` |
| 无业务语义 UI  | `components/ui`                                                            |
| 工作流步骤表单 | `features/workflow/workflows/*`                                            |

**页面模板：** `PageShell` → `DataSection`（loading / error / empty）→ 领域内容；Error 用 `ErrorState` + hook 的 `refresh`。

### 2.2 技术选型（勿换）

React + Vite、TanStack Query、React Router、Zustand（仅 workflow）、Radix/shadcn、Vitest + Playwright。**不引入** MSW（已移除）、Redux、Next.js。测试：Vitest 用 `createMockApis()`；E2E/dev 用真 backend。

**可补充（见 plan §4）：** Zod 扩大覆盖面、OpenAPI/orval、`@tanstack/react-virtual`（大表按需）、`eslint-plugin-boundaries`。

---

## 3. 路由

[`config/routes.ts`](../apps/frontend/src/config/routes.ts) 以 **`ROUTE_DEFINITIONS`** 为唯一源，派生 `ROUTES`、`APP_ROUTES` 等。

当前 **17** 业务页：dashboard（cost、usage）、org（3）、budget（2：index、alerts）、keys（4）、models（2）、wallet（1）、audit（2）。`/billing` 重定向至 `/wallet`。

新增页面：在 `ROUTE_DEFINITIONS` 加一条 → `features/{domain}/hooks/use-{page}-page.ts` + shell → `routes/{domain}/{page}.tsx` 薄入口。

---

## 4. API 层与页面架构

- `api/client.ts`：`request()`、`ApiError`、`buildQuery()`；`credentials: 'include'`
- `app-apis.ts`：`AppApis` + `defaultApis`（**16** 命名空间；仍缺 `platformApi`）
- 生产 `AppProviders`（`components/layout/app-providers.tsx`）注入 `defaultApis`；测试 `createMockApis()`

**薄页面 → 页面 Hook → 展示组件**：Hook 用 `useApis()`、`useInjectedQuery` + `queryKeys`；组件 props 受控。

**Workflow：** `features/workflow/` Zustand 侧滑栈；`workflows/{name}.tsx` + `definitions/` 注册。

**SaaS 部分接入：** 企业面 `authApi`（login/logout）、`billingApi`（wallet/recharge/confirm）与 `/wallet` 页已接入；`accept-invite` 与 `/platform/*` 仍无前端。见 §5.9.6、[Roadmap.md](./Roadmap.md)。

---

## 5. API 契约

### 5.0 调用架构

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
| `app-apis.ts`           | `api/app-apis.ts`                      | `AppApis` 接口与 `defaultApis` 聚合（16 个命名空间）                |
| `api-context.ts`        | `api/api-context.ts`                   | `ApiContext` React Context                                          |
| `context.tsx`           | `api/context.tsx`                      | `ApiProvider` 注入                                                  |
| `use-apis.ts`           | `api/use-apis.ts`                      | `useApis()`、`useInjectedApis()`                                    |
| `use-injected-query.ts` | `features/query/use-injected-query.ts` | 基于 TanStack Query 的 `useInjectedQuery()`                         |
| `query-keys.ts`         | `features/query/query-keys.ts`         | 各域 query key 工厂                                                 |
| `{domain}.ts`           | `api/{domain}.ts`                      | 各资源 HTTP 方法                                                    |

**依赖注入：** 生产环境 `AppProviders` 注入 `defaultApis`；页面 Hook 支持 `injectedApis` 参数供测试覆盖；测试通过 `createMockApis()` 注入 mock API。

**数据获取：** 读操作逐步迁移至 `useInjectedQuery` + `queryKeys`；写操作仍在事件处理函数中直接调用 `*Api` 方法，成功后通过 `queryClient.invalidateQueries` 或 `useWorkflowRefresh` 刷新缓存。

### 5.0.1 领域数据约定（Keys / Models）

| 决策                     | 约定                                                                          |
| ------------------------ | ----------------------------------------------------------------------------- |
| 列表 enrich              | 扩展现有 `GET /keys/platform`、`GET /models`；**不**新增平行 `/enriched` 端点 |
| `platform_keys` 推导字段 | **不入库**；`platform_key_enrich.go` domain join；改名后下次 GET 自动反映     |
| 审批「我的申请」         | `memberId` 查询参数；`tab` 仅表状态维度                                       |
| `models.visibility`      | 可编辑、展示；运行时与 allowlist 合并校验属 [plan.md](./plan.md) §7           |
| 发布                     | 前后端同发；DB 迁移 additive only                                             |

**`platform_keys` 字段分层：** 持久化 `member_id` / `budget_group_id` / `app_name`；响应 enrich `member_name` / `budget_group_name` / `type` / `department_*` / `project_name`（仅 JSON）；运行面 `relay_mappings.department_id` 独立分层。

**`models` 表扩展列：** `model_type`（`builtin`/`custom`）、`description`、`visibility`、`endpoint`（custom 部署地址）。生产迁移 SQL 见 [plan.md](./plan.md) §5。

---

### 5.1 通用约定

#### 5.1.1 Base URL

| 常量            | 值               | 定义                                                  |
| --------------- | ---------------- | ----------------------------------------------------- |
| `API_BASE_PATH` | `{BASE_URL}/api` | [`config/app.ts`](../apps/frontend/src/config/app.ts) |

本地开发与 nginx 同域部署均为 `/api`。

开发环境通过 Vite dev/preview 将 `{BASE_URL}/api` 同域反代到 Go（默认 `http://127.0.0.1:8080`，可用 `VITE_API_PROXY_TARGET` 覆盖）。生产使用 nginx 等同域反代，见 [`deploy/nginx.conf.example`](../deploy/nginx.conf.example)。

#### 5.1.2 请求头

| Header          | 必填       | 说明                                                                                                                                    |
| --------------- | ---------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `Accept`        | 是         | `application/json`（`client.request` 默认注入）                                                                                         |
| `Content-Type`  | 有 body 时 | `application/json`（`client.request` 默认注入）                                                                                         |
| `Cookie`        | Session    | `credentials: 'include'` 自动携带；企业面 `tokenjoy_session_member`；平台面 `tokenjoy_platform_session`（`SUPPORT_SAAS=true`，见 §5.9） |
| `Authorization` | 可选       | `Bearer {token}`（生产网关）                                                                                                            |

#### 5.1.3 成功响应

- HTTP 2xx
- 返回 JSON，结构与 `api/types/` 一致
- `void` 类型端点：TypeScript 侧为 `void`；**注意** `request()` 始终调用 `res.json()`，后端对无 body 端点应返回 `{}` 或 `null`，避免空 body 导致解析失败

#### 5.1.4 错误响应

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

#### 5.1.5 分页

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

#### 5.1.6 查询参数构建

`buildQuery()` 会跳过 `undefined`、`null`、空字符串。布尔值序列化为 `"true"` / `"false"`。

#### 5.1.7 命名与参数约定

跨层术语见 [Backend-存储.md](./Backend-存储.md) §10。HTTP 契约要点：

| 约定                                                  | 说明                                                                       |
| ----------------------------------------------------- | -------------------------------------------------------------------------- |
| `departmentId`（JSON）                                | 组织节点 ID（`org_nodes.id`）；与存储列 `department_id` 同语义             |
| `departmentId`（budget path / JSON）                  | org_node ID；`/budget/departments/:departmentId` 与响应字段同名            |
| `deptId`（path / query，非 budget）                   | dashboard `/cost/departments/{deptId}/...`、`?deptId=` 等                  |
| 前端 API 客户端参数                                   | budget 域用 `departmentId`；dashboard/models query 映射为 `deptId`         |
| `PUT /budget/departments/:departmentId`               | 更新部门节点预算                                                           |
| `GET /budget/departments/:departmentId/member-quotas` | 该部门下成员配额列表                                                       |
| `GET /dashboard/usage/teams`                          | 产品「团队用量」；响应字段仍为 `departmentId` / `departmentName`           |
| `RoutingRule.id` 与 `nodeId`                          | 同值；新代码优先读 `nodeId`；`PUT /models/routing/:id` 中 `:id` = `nodeId` |

---

### 5.2 鉴权与权限（目标态）

**方案 B**：签名 Session JWT（仅 identity）+ 服务端 PDP；UI **一次** `GET /session` → Context。破坏性实现规格见 [权限管理.md](./权限管理.md)。

#### 5.2.1 Session 端点

| 调用方式                  | 说明                                                                |
| ------------------------- | ------------------------------------------------------------------- |
| `sessionApi.getCurrent()` | `GET /session`；凭 HttpOnly Cookie（JWT）或 `Authorization: Bearer` |

响应经 `SessionContextSchema`（Zod）校验，**必含** `authzRevision`。未鉴权 → `401`；成员不存在 → `404`。`SessionGate` 处理加载失败。

**删除（目标态）**：`/login` 选择成员写裸 `member_id` cookie；改为 `POST /auth/login`。

#### 5.2.2 鉴权策略

| 范围         | 要求                                                                |
| ------------ | ------------------------------------------------------------------- |
| 全部业务 API | Session JWT + 对应读/写 capability                                  |
| 公开         | `POST /auth/login`、`POST /auth/logout`、`POST /auth/accept-invite` |

**删除**：`APP_PROFILE=demo` 下 GET 免 Session。

#### 5.2.3 权限驱动 UI

- `permissions[]` **仅**来自 `GET /session`；禁止前端 `resolveMemberPermissions` 生产路径。
- `PermissionGate` / `usePermissions()`：**零** per-component session 请求。
- 失效：`refreshSession()`（PAP mutation、focus、broadcast、403）；比对 `authzRevision`。

#### 5.2.5 SaaS 双 Session

| 面     | Cookie                      | 登录入口                     | 说明                          |
| ------ | --------------------------- | ---------------------------- | ----------------------------- |
| 企业面 | `tokenjoy_session_member`   | `POST /auth/login`、邀请激活 | JWT；`companyId` 来自 Session |
| 平台面 | `tokenjoy_platform_session` | `POST /platform/auth/login`  | 独立 JWT；不走企业 RBAC       |

#### 5.2.4 权限 Key

由 **`manifest.json` 生成** [`permission-keys.ts`](../apps/frontend/src/lib/permission-keys.ts)（目标态）。完整表见 [权限管理.md](./权限管理.md) §14.1。

---

### 5.3 共享类型

定义于 [`api/types/common.ts`](../apps/frontend/src/api/types/common.ts)。

### Paginated\<T\>

见 §5.1.5。

### SessionContext

| 字段            | 类型       | 说明                                  |
| --------------- | ---------- | ------------------------------------- |
| `member`        | `Member`   | 当前成员                              |
| `permissions`   | `string[]` | PDP 输出的 capability 列表            |
| `readOnly`      | `boolean`  | 无写 capability 时为 `true`           |
| `companyId`     | `number`   | 当前成员所属企业 ID                   |
| `authzRevision` | `number`   | 租户授权版本；UI stale 检测（目标态） |

---

### 5.4 端点清单

路径均相对于 `API_BASE_PATH`（`/api`）。共 **82** 个企业面业务端点（session → audit，不含 §5.9 的 auth/billing/platform 扩展）。类型详情见 §5.5。

#### 5.4.1 Session

客户端：[`sessionApi`](../apps/frontend/src/api/session.ts)

| 方法 | 路径       | 查询 / Body | 响应             | 说明                                   |
| ---- | ---------- | ----------- | ---------------- | -------------------------------------- |
| GET  | `/session` | —           | `SessionContext` | 签名 JWT Cookie / Bearer；未鉴权 → 401 |

---

#### 5.4.2 Org（组织管理）

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

#### 5.4.3 Budget（预算管理）

客户端：[`budgetApi`](../apps/frontend/src/api/budget.ts)

| 方法   | 路径                                              | Body / 查询                                      | 响应                  | 备注                              |
| ------ | ------------------------------------------------- | ------------------------------------------------ | --------------------- | --------------------------------- |
| GET    | `/budget/tree`                                    | query: `period?`                                 | `BudgetNode[]`        |                                   |
| PUT    | `/budget/departments/:departmentId`               | `{ budget, reservedPool? }`                      | `BudgetNode`          |                                   |
| GET    | `/budget/departments/:departmentId/member-quotas` | —                                                | `MemberBudgetQuota[]` |                                   |
| PUT    | `/budget/members/:memberId`                       | `UpdateMemberQuotaInput`                         | `MemberBudgetQuota`   | 部门内超卖 / 低于已分配 Key → 422 |
| GET    | `/budget/groups`                                  | —                                                | `BudgetGroup[]`       |                                   |
| POST   | `/budget/groups`                                  | `Omit<BudgetGroup, 'id' \| 'consumed'>`          | `BudgetGroup`         |                                   |
| PUT    | `/budget/groups/:id`                              | `Partial<Omit<BudgetGroup, 'id' \| 'consumed'>>` | `BudgetGroup`         |                                   |
| DELETE | `/budget/groups/:id`                              | —                                                | `void`                |                                   |
| GET    | `/budget/overrun-policy`                          | —                                                | `OverrunPolicyConfig` |                                   |
| PUT    | `/budget/overrun-policy`                          | `OverrunPolicyConfig`                            | `OverrunPolicyConfig` |                                   |
| GET    | `/budget/alerts`                                  | —                                                | `AlertRule[]`         |                                   |
| POST   | `/budget/alerts`                                  | `Omit<AlertRule, 'id'>`                          | `AlertRule`           |                                   |
| PUT    | `/budget/alerts/:id`                              | `Partial<AlertRule>`                             | `AlertRule`           |                                   |
| DELETE | `/budget/alerts/:id`                              | —                                                | `void`                |                                   |

---

#### 5.4.4 Keys（API Key 管理）

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

| 方法   | 路径                           | Body / 查询                                                                          | 响应                     | 备注                                                                 |
| ------ | ------------------------------ | ------------------------------------------------------------------------------------ | ------------------------ | -------------------------------------------------------------------- |
| GET    | `/keys/platform`               | query: `page?`, `pageSize?`, `memberId?`, `budgetGroupId?`, `departmentId?`, `type?` | `Paginated<PlatformKey>` | 服务端筛选 + enrich；`type`: `member` \| `project`                   |
| POST   | `/keys/platform`               | `{ name, memberId?, budgetGroupId?, appName?, quota, modelWhitelist }`               | `PlatformKey`            | 个人 Key 缺 `memberId` → 400；Group Key 校验组剩余额度；白名单 → 422 |
| PUT    | `/keys/platform/:id`           | `{ name?, quota?, modelWhitelist? }`                                                 | `PlatformKey`            | 额度 / 白名单校验 → 422                                              |
| PUT    | `/keys/platform/:id/toggle`    | `{ enabled }`                                                                        | `PlatformKey`            |                                                                      |
| POST   | `/keys/platform/:id/rotate`    | —                                                                                    | `PlatformKey`            | 响应含 `fullKey`                                                     |
| PUT    | `/keys/platform/:id/revoke`    | —                                                                                    | `void`                   |                                                                      |
| DELETE | `/keys/platform/:id`           | —                                                                                    | `void`                   |                                                                      |
| GET    | `/keys/platform/quota-summary` | query: `memberId`                                                                    | `MemberQuotaSummary`     |                                                                      |

#### 审批 `approvalApi`

| 方法 | 路径                              | Body / 查询                                                   | 响应                                      | 备注                                |
| ---- | --------------------------------- | ------------------------------------------------------------- | ----------------------------------------- | ----------------------------------- |
| GET  | `/keys/approvals`                 | query: `tab?`, `memberId?`                                    | `KeyApproval[]`                           | `tab`: `pending` \| `mine` \| `all` |
| POST | `/keys/approvals`                 | `{ type, reason, requestedQuota, requestedModels, memberId }` | `KeyApproval`                             | 白名单校验 → 422                    |
| PUT  | `/keys/approvals/:id/approve`     | —                                                             | `void`                                    | 预留池不足 → 422                    |
| PUT  | `/keys/approvals/:id/reject`      | `{ reason? }`                                                 | `void`                                    |                                     |
| GET  | `/keys/approvals/:id/quota-check` | —                                                             | `{ sufficient, reservedPool, requested }` |                                     |

---

#### 5.4.5 Models（模型与路由）

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

#### 5.4.6 Dashboard（数据看板）

客户端：[`dashboardApi`](../apps/frontend/src/api/dashboard.ts)

**只读约束：** 本节 **全部端点为 `GET`**，查询用量/成本，**无副作用**（不写库、不触发 ingest、不修改预算）。实现架构见 [Backend-架构.md](./Backend-架构.md) §8。

成本类接口共享查询参数 `CostQueryParams`（`period` 默认 `current_month`）：

| 字段          | 含义                                                                         |
| ------------- | ---------------------------------------------------------------------------- |
| `period`      | `current_month` \| `last_month` \| `last_7_days` \| `custom`                 |
| `startDate`   | `period=custom` 时必填，ISO 日期（按响应 `timezone` 解释）                   |
| `endDate`     | `period=custom` 时必填，ISO 日期                                             |
| `granularity` | 趋势粒度：`day` \| `hour` \| `week` \| `month`（`minute` 仅 `usage/series`） |

**数据源约定**

- `cost/*`、`usage/models`、`usage/teams` 的 **consumed / cost** 均来自 **`usage_buckets` 周期聚合**（不读 `org_nodes.consumed`）。
- `usage/teams` 的 **quota** 来自 **`org_nodes`** 预算树。
- 聚合/展示时区默认 **`Asia/Shanghai`**（IANA）；响应 `timezone` 字段返回实际使用值；企业可配置覆盖。
- `week` / `month` 由服务端对 buckets 做 `date_trunc('week' \| 'month', …)`，不走 `usage/series`。

**用量时间序列** — 统一 `day` / `hour` / `minute` 查询形状：

| 方法 | 路径                      | 查询               | 响应                  |
| ---- | ------------------------- | ------------------ | --------------------- |
| GET  | `/dashboard/usage/series` | `UsageSeriesQuery` | `UsageSeriesResponse` |

`UsageSeriesQuery`：`granularity`（`day` \| `hour` \| `minute`，必填）、`start`、`end`（ISO8601 或 `YYYY-MM-DD`）、`groupBy?`（`none` \| `department` \| `member` \| `model`，**单选**）、`departmentId?`、`memberId?`

| 路径           | 数据源                  | 说明                                                                                                              |
| -------------- | ----------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `day` / `hour` | `usage_buckets`         | `source: "buckets"`，`approximate: false`，`mappingAsOf: "ingest_time"`                                           |
| `minute`       | `usage_ledger` 分钟聚合 | `source: "ledger"`，`approximate: false`，`mappingAsOf: "ingest_time"`；**最大窗口 3h**；禁止与 hour/day 混合环比 |

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

#### 5.4.7 Audit（审计日志）

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

**调用审计：** `GET /audit/calls` 只读 `usage_ledger`；`keyword` 匹配 `previewSnippet` 等字段。账本仅存 **input** 截断 snippet，不存 output 原文；`outputTokens` 仅为用量计数。不查 NewAPI `logs`；不提供全文 content 接口（首版）。详见 [Backend-存储.md](./Backend-存储.md) §6、[Backend-预算.md](./Backend-预算.md) §2。

---

### 5.5 类型参考

完整定义以源码为准。以下列出各域核心类型字段。

#### 5.5.1 Org — [`types/org.ts`](../apps/frontend/src/api/types/org.ts)

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

#### 5.5.2 Budget — [`types/budget.ts`](../apps/frontend/src/api/types/budget.ts)

**BudgetNode：** `id`, `name`, `parentId`, `budget`, `consumed`, `reservedPool?`, `children?`, `period`

**MemberBudgetQuota：** `memberId`, `memberName`, `departmentId`, `personalQuota`, `allocated`, `used`

**UpdateMemberQuotaInput：** `personalQuota`

**BudgetGroup：** `id`, `name`, `budget`, `consumed`, `memberIds`, `departmentIds`

**OverrunPolicyConfig：** `thresholds`, `notifyEmail`, `notifyPhone`, `notifyIm`, `blockMessage`

**AlertRule：** `id`, `nodeId`, `nodeName`, `thresholds`, `notifyRoleIds`, `enabled`

**CreateModelInput：** `name`, `displayName`, `baseUrl`, `apiKey`, `inputPrice`, `outputPrice`

**ResolvedWhitelist：** `inherited`, `allowedModels`, `parentCount`

#### 5.5.3 Keys — [`types/keys.ts`](../apps/frontend/src/api/types/keys.ts)

**ProviderType：** `openai` \| `anthropic` \| `deepseek` \| `qwen` \| `custom`

**KeyStatus：** `active` \| `disabled` \| `expired` \| `error`

**ProviderKey：** `id`, `provider`, `name`, `keyPrefix`, `status`, `balance`, `lastUsed`, `createdAt`, `rotateEnabled`

**PlatformKey：** `id`, `name`, `keyPrefix`, `fullKey?`, `memberId`, `memberName`†, `appName`, `budgetGroupId`, `budgetGroupName`†, `type`†, `departmentId`†, `departmentName`†, `projectName`†, `status`, `quota`, `used`, `modelWhitelist`, `createdAt`, `expiresAt`  
† 服务端 enrich 推导，不入库 `platform_keys`（见 §5.0.1）

**ApprovalType：** `key` \| `quota` · **ApprovalStatus：** `pending` \| `approved` \| `rejected`

**KeyApproval：** `id`, `type`, `applicant`, `applicantId`, `department`, `reason`, `requestedQuota`, `requestedModels`, `status`, `approver`, `rejectReason?`, `createdAt`, `resolvedAt`

**MemberQuotaSummary：** `totalQuota`, `used`, `remaining`, `reservedPool`

#### 5.5.4 Models — [`types/models.ts`](../apps/frontend/src/api/types/models.ts)

**ModelInfo：** `id`, `provider`, `name`, `displayName`, `inputPrice`, `outputPrice`, `maxContext`, `enabled`, `capabilities`

**RoutingRule：** `id`（= `nodeId`）, `nodeId`, `nodeName`, `allowedModels`, `defaultModel`, `fallbackModel`, `inherited`

#### 5.5.5 Dashboard — [`types/dashboard.ts`](../apps/frontend/src/api/types/dashboard.ts)

**CostPeriod：** `current_month` \| `last_month` \| `last_7_days` \| `custom`

**CostGranularity：** `day` \| `hour` \| `week` \| `month`

**UsageGranularity：** `day` \| `hour` \| `minute`

**UsageSeriesGroupBy：** `none` \| `department` \| `member` \| `model`

**UsageSeriesSource：** `buckets` \| `ledger`

**UsageMappingAsOf：** `ingest_time` \| `query_time`

**CostQueryParams：** `period?`, `startDate?`, `endDate?`, `granularity?`

**UsageSeriesQuery：** `granularity`, `start`, `end`, `groupBy?`, `departmentId?`, `memberId?`

**UsageSeriesPoint：** `bucket`（ISO8601 时间桶起点，含时区偏移）, `departmentId?`, `memberId?`, `model?`, `costCny`, `callCount`, `inputTokens`, `outputTokens`

**UsageSeriesResponse：** `granularity`, `source`, `timezone`（默认 `Asia/Shanghai`）, `approximate`, `mappingAsOf`, `unmappedCount?`, `truncated?`, `points: UsageSeriesPoint[]`

| 字段            | buckets 路径    | minute 路径     |
| --------------- | --------------- | --------------- |
| `source`        | `"buckets"`     | `"ledger"`      |
| `approximate`   | `false`         | `false`         |
| `mappingAsOf`   | `"ingest_time"` | `"ingest_time"` |
| `unmappedCount` | 省略            | 省略            |
| `truncated`     | 省略或 `false`  | 省略或 `false`  |

当前实现：`inputTokens` / `outputTokens` 恒为 `0`（webhook 未带 token 字段）。

**CostSummary：** `totalCost`, `totalCostMom`, `totalTokens`, `totalRequests`, `totalRequestsMom`, `avgCostPerRequest`, `avgCostPerRequestMom`, `avgCostPerMember`, `avgCostPerMemberMom`

**DepartmentCost：** `departmentId`, `departmentName`, `cost`, `percentage`, `hasChildren?`

**DepartmentCostMember：** `memberId`, `memberName`, `cost`, `requests`, `tokens`

**DailyCost：** `date`, `cost`, `tokens`, `requests`

**TopConsumer：** `memberId`, `memberName`, `department`, `cost`, `tokens`, `requests`

**ModelUsage：** `modelId`, `modelName`, `provider`, `requests`, `tokens`, `cost`, `percentage`

**TeamUsage：** `departmentId`, `departmentName`, `quota`, `consumed`, `memberCount`, `topModel`

#### 5.5.6 Audit — [`types/audit.ts`](../apps/frontend/src/api/types/audit.ts)

**AuditAction：** `key_create` \| `key_disable` \| `key_rotate` \| `budget_change` \| `budget_approve` \| `permission_change` \| `role_assign` \| `model_whitelist_change` \| `member_add` \| `member_remove` \| `org_structure_change`

**OperationLog：** `id`, `action`, `operator`, `operatorId`, `target`, `detail`, `ip`, `createdAt`

**CallLog：** `id`, `caller`, `callerId`, `callerType`（`member` \| `platform_key`）, `model`, `provider`, `inputTokens`, `outputTokens`, `latencyMs`, `status`, `cost`, `createdAt`, `previewSnippet`

- `previewSnippet`：仅 **input** 截断 ~200 字；`contentRetentionEnabled=false` 或 Webhook 无 `input` 时为空串
- `outputTokens`：completion token **计数**，非 output 正文
- 首版**不提供** input 全文或 output 原文；展开行仅展示 `previewSnippet`

**AuditSettings：** `contentRetentionEnabled` — `false` 时不写 `previewSnippet`

---

### 5.6 AppApis 聚合

[`app-apis.ts`](../apps/frontend/src/api/app-apis.ts) 中 `defaultApis` 包含 **16** 个命名空间（仍缺 `platformApi`，见 §5.9）：

`sessionApi`, `authApi`, `billingApi`, `dataSourceApi`, `syncApi`, `departmentApi`, `memberApi`, `roleApi`, `budgetApi`, `providerKeyApi`, `platformKeyApi`, `approvalApi`, `modelApi`, `routingApi`, `dashboardApi`, `auditApi`

所有 HTTP 调用均经 `client.request()`，无其他 `fetch('/api/...')` 直连。

---

### 5.7 开发与部署

#### 5.7.1 环境变量

| 变量                    | 说明                                                     |
| ----------------------- | -------------------------------------------------------- |
| `VITE_API_PROXY_TARGET` | 可选；覆盖 Vite 反代目标（默认 `http://127.0.0.1:8080`） |
| `BASE_URL`              | Vite 应用根路径；影响 `API_BASE_PATH`（默认 `/`）        |

根目录 `pnpm start` 并发启动 backend + frontend；[`apps/frontend/.env.development`](../apps/frontend/.env.development) 默认配置代理目标。

#### 5.7.2 鉴权（目标态）

- **企业面**：[`/login`](../apps/frontend/src/routes/auth/login.tsx) 表单 → `POST /auth/login` → 后端 `Set-Cookie`（JWT）；登出 `authApi.logout()` → `POST /auth/logout`
- **平台面（SaaS）**：`POST /platform/auth/login` → 平台 JWT Cookie（前端未接入）
- **邀请激活**：`POST /auth/accept-invite` 签发企业 JWT（前端未接入独立页）
- 401 → `AuthUnauthorizedBridge` 跳转 `/login`（或 `/platform/login`）

**删除**：Dev 成员选择器直写裸 `member_id`；见 [权限管理.md](./权限管理.md) §2。

#### 5.7.3 联调验收

1. 启动 `pnpm start` 或分别启动 backend / frontend
2. 登录后逐域抽查：session → org → budget → keys → models → dashboard → billing → audit
3. 看板 `/dashboard/cost` 切换 `granularity`（day/week/month）应即时刷新

---

### 5.8 变更检查清单

新增或修改 API 时，同步更新：

- [ ] `api/{domain}.ts` — HTTP 方法
- [ ] `api/types/{domain}.ts` — 请求 / 响应类型
- [ ] `api/app-apis.ts` — 若新增域 API 对象
- [ ] `api/schemas/` — 若响应需运行时校验（如 Session）
- [ ] `features/query/query-keys.ts` — 若读操作使用 React Query
- [ ] 本文档 §5.4 端点清单（SaaS 见 §5.9）
- [ ] 后端 handler + domain service（[`apps/backend/`](../apps/backend/)）
- [ ] 页面 Hook / 测试 — 按需补充

---

### 5.9 SaaS 多企业

产品模型：**企业（Company）** = 一家公司；**成员（User）** = 企业内员工。计费双轴（企业钱包 + 部门 budget）、Gateway、平台鉴权详见 [Backend-预算.md](./Backend-预算.md) §1、[Backend.md](./Backend.md) §2。

**共 11 个端点**（路径均相对于 `API_BASE_PATH`）。

#### 5.9.1 企业面：认证

客户端：[`auth.ts`](../apps/frontend/src/api/auth.ts)。`login` / `logout` 已接入 `AppApis`；`accept-invite` 尚无前端封装与路由。

| 方法 | 路径                  | Body                         | 响应                   | 说明                                                     |
| ---- | --------------------- | ---------------------------- | ---------------------- | -------------------------------------------------------- |
| POST | `/auth/login`         | `{ email, password }`        | `{ memberId: string }` | 签发企业 JWT Cookie                                      |
| POST | `/auth/logout`        | —                            | `void`                 | 清 Cookie；MVP 不维护服务端吊销集                        |
| POST | `/auth/accept-invite` | `{ token, password, name? }` | `SessionContext`       | 邀请激活；签发 JWT Cookie；`token` 一次性、默认 7 天有效 |

无需 Session；`accept-invite` 成功后与 `GET /session` 结构一致（含 `companyId`）。

#### 5.9.2 企业面：计费

客户端：[`billing.ts`](../apps/frontend/src/api/billing.ts)。`getWallet` / `recharge` / `confirmRecharge` 已接入 `use-wallet-page`。

| 方法 | 路径                             | Body / 查询                  | 响应            | 权限               | 说明                              |
| ---- | -------------------------------- | ---------------------------- | --------------- | ------------------ | --------------------------------- |
| GET  | `/billing/wallet`                | —                            | `WalletSummary` | `billing:read`     | 读 NewAPI 企业钱包                |
| POST | `/billing/recharge`              | `{ amount, idempotencyKey }` | `RechargeOrder` | `billing:recharge` | 创建 `pending` 订单；HTTP **202** |
| POST | `/billing/recharge/{id}/confirm` | —                            | `void`          | `billing:recharge` | 确认支付并入账（demo / 回调模拟） |

**`WalletSummary`**

| 字段          | 类型     | 说明                                 |
| ------------- | -------- | ------------------------------------ |
| `balance`     | `number` | 企业钱包余额（CNY）                  |
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

#### 5.9.3 平台面：认证

后端已实现；前端尚无 `platformApi` / 平台路由。

| 方法 | 路径                   | Body                  | 响应              | 说明                             |
| ---- | ---------------------- | --------------------- | ----------------- | -------------------------------- |
| POST | `/platform/auth/login` | `{ email, password }` | `PlatformSession` | 写入 `tokenjoy_platform_session` |

**`PlatformSession`**

| 字段       | 类型                            | 说明         |
| ---------- | ------------------------------- | ------------ |
| `operator` | `{ id: string; email: string }` | 平台运营账号 |

其余 `/platform/*` 须携带平台 Session；`SUPPORT_SAAS=false` 时返回 404。

#### 5.9.4 平台面：企业与 Channel

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

#### 5.9.5 既有端点的 SaaS 行为差异

| 端点 / 域                      | 私有化          | SaaS                                                            |
| ------------------------------ | --------------- | --------------------------------------------------------------- |
| `GET /session`                 | `companyId = 1` | `companyId` = 成员所属企业                                      |
| `provider-keys` 写             | 允许            | **403**                                                         |
| `POST /keys/platform`          | 不变            | 不变；后端 Token 绑 `newapi_wallet_user_id` + `platform_shared` |
| 组织 / 预算 / 密钥读写在企业面 | 单企业          | 按 Session `companyId` 隔离                                     |
| 成员邀请（企业内）             | 支持            | `POST /org/members/invite`                                      |

#### 5.9.6 前端接入现状

| 项                   | 后端                          | 前端 `AppApis`                            | 控制台页面                         |
| -------------------- | ----------------------------- | ----------------------------------------- | ---------------------------------- |
| 企业面 §5.4 域 API   | 已实现                        | 已接入（16 命名空间，缺 `platformApi`）   | 17 业务页                          |
| `auth/login`         | 已实现                        | `authApi.login`                           | `/login`                           |
| `auth/logout`        | 已实现                        | `authApi.logout`                          | —                                  |
| `auth/accept-invite` | 已实现                        | 未接入                                    | 无 `/invite/accept`                |
| `billing/wallet`     | 已实现                        | `billingApi.getWallet`                    | `/wallet`                          |
| `billing/recharge`   | 已实现                        | `billingApi.recharge` + `confirmRecharge` | `/wallet`（充值 create → confirm） |
| `platform/*`         | 已实现（`SUPPORT_SAAS=true`） | 未接入                                    | 无 `/platform/login`               |
| `billing:*` 权限     | 已挂 Authz                    | `permission-keys.ts` 已含                 | `PermissionGate` 已用于 `/wallet`  |

> **类型对齐：** 前端 `WalletView` 已与后端 `WalletSummary` 对齐（`balance` / `allocatable`）。`totalConsumed` / `totalRequests` 为半真聚合，真实现见 [plan.md](./plan.md) §2。

后端详案：[Backend.md](./Backend.md) §2。NewAPI 部署：[Backend.md](./Backend.md) §4。

---

## 6. 本地联调

```bash
pnpm install && pnpm start
```

| 服务 | 地址                  |
| ---- | --------------------- |
| 前端 | http://localhost:5173 |
| 后端 | http://localhost:8080 |

1. `/login` 用种子账号登录（见 `权限管理.md` WP-2.6）→ JWT Cookie
2. 空库自动 `seed.ApplyTables`；看板锚定 `DEMO_TODAY=2026-06-19`
3. 重置：`pnpm docker:reset && pnpm start`

| 变量                    | 说明                                   |
| ----------------------- | -------------------------------------- |
| `VITE_API_PROXY_TARGET` | 反代目标，默认 `http://127.0.0.1:8080` |
| `DATABASE_URL`          | 后端 Postgres                          |

可选 Relay：`pnpm start:relay` + `NEW_API_ENABLED=true`。验收：`pnpm verify`、`pnpm test:e2e`。

生产同域：`deploy/nginx.conf.example`（`/api/` 须在 SPA fallback 之前）。

---

## 7. 测试与 PR 自检

`tests/setup.ts`、`createMockApis()`、`renderHookWithProviders()`；不依赖 backend 进程。

- [ ] 新页面只改 `ROUTE_DEFINITIONS` 一条
- [ ] 页面从 `use-*-page` 取数
- [ ] 新 API：§5.4 契约 + `api/` + `queryKeys` + 后端 handler
- [ ] `pnpm lint` 与 `pnpm test` 通过

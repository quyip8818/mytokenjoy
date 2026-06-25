# Frontend API 契约

本文档描述 TokenJoy 前端 `src/api/` 层调用的 REST 接口契约，供 MSW mock 与真实后端实现对齐。

**权威来源（按优先级）：**

1. 类型定义 — [`apps/frontend/src/api/types/`](../apps/frontend/src/api/types/)
2. HTTP 客户端 — [`apps/frontend/src/api/{domain}.ts`](../apps/frontend/src/api/)
3. MSW handlers — [`apps/frontend/src/mocks/handlers/`](../apps/frontend/src/mocks/handlers/)

架构与分层约定见 [Frontend-代码结构.md](./Frontend-代码结构.md)。

---

## 1. 调用架构

```
UI / 页面 Hook
  └─ useApis() → AppApis（依赖注入）
       └─ {domain}Api.{method}()
            └─ client.request() → fetch({API_BASE_PATH}{path})
                 ├─ USE_MOCKS=true  → MSW domainHandlers + fallbackHandlers
                 └─ USE_MOCKS=false → 真实后端（或 Vite proxy）
```

| 模块          | 路径                         | 职责                                                                 |
| ------------- | ---------------------------- | -------------------------------------------------------------------- |
| `client.ts`   | `api/client.ts`              | `request()`、`ApiError`、`buildQuery()`、`setDemoMemberIdProvider()` |
| `app-apis.ts` | `api/app-apis.ts`            | `AppApis` 接口与 `defaultApis` 聚合                                  |
| `context.tsx` | `api/context.tsx`            | `ApiProvider` 注入                                                   |
| `use-apis.ts` | `api/use-apis.ts`            | React 消费入口                                                       |
| `{domain}.ts` | `api/{domain}.ts`            | 各资源 HTTP 方法                                                     |
| handlers      | `mocks/handlers/{domain}.ts` | 与 api 同域的 mock 实现                                              |

**依赖注入：** 生产环境 `AdminLayout` 注入 `defaultApis`；页面 Hook 支持 `injectedApis` 参数供测试覆盖；非 React 代码（如 `createDemoRoleStore`）通过构造函数注入 `sessionApi` 等。

---

## 2. 通用约定

### 2.1 Base URL

| 常量            | 值               | 定义                                                  |
| --------------- | ---------------- | ----------------------------------------------------- |
| `API_BASE_PATH` | `{BASE_URL}/api` | [`config/app.ts`](../apps/frontend/src/config/app.ts) |

本地开发通常为 `/api`；GitHub Pages 子路径部署时为 `/{repo}/api`。

### 2.2 请求头

| Header             | 必填       | 说明                                                                                        |
| ------------------ | ---------- | ------------------------------------------------------------------------------------------- |
| `Content-Type`     | 有 body 时 | `application/json`（`client.request` 默认注入）                                             |
| `X-Demo-Member-Id` | Demo 模式  | 当前 Demo 成员 ID，由 `setDemoMemberIdProvider` 注入；真实后端应替换为 session / token 鉴权 |

### 2.3 成功响应

- HTTP 2xx
- `void` 类型端点可返回空 body 或 `null`
- 其余端点返回 JSON，结构与 `api/types/` 一致

### 2.4 错误响应

HTTP 非 2xx 时，body 应包含：

```json
{ "message": "错误描述" }
```

前端 `ApiError`（[`client.ts`](../apps/frontend/src/api/client.ts)）读取 `message` 字段，无则回退 `statusText`。

**常见状态码：**

| 状态码 | 场景                                                                 |
| ------ | -------------------------------------------------------------------- |
| `400`  | 缺少必填参数（如 session 缺 `memberId`、platform key 缺 `memberId`） |
| `404`  | 资源不存在                                                           |
| `422`  | 业务校验失败（额度不足、模型不在部门白名单等）                       |
| `501`  | Demo 模式下未实现的路径（`fallbackHandlers` 返回）                   |

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

## 3. 共享类型

定义于 [`api/types/common.ts`](../apps/frontend/src/api/types/common.ts)。

### Paginated\<T\>

见 §2.5。

### SessionContext

| 字段          | 类型       | 说明                                         |
| ------------- | ---------- | -------------------------------------------- |
| `member`      | `Member`   | 当前成员                                     |
| `permissions` | `string[]` | 权限 key 列表（见 `lib/permission-keys.ts`） |
| `readOnly`    | `boolean`  | 无写权限时为 `true`                          |

---

## 4. 端点清单

路径均相对于 `API_BASE_PATH`。类型详情见 §5。

### 4.1 Session

客户端：[`sessionApi`](../apps/frontend/src/api/session.ts) · Handler：[`session.ts`](../apps/frontend/src/mocks/handlers/session.ts)

| 方法 | 路径       | 查询 / Body        | 响应             | 说明                                  |
| ---- | ---------- | ------------------ | ---------------- | ------------------------------------- |
| GET  | `/session` | Demo：`memberId`（必填）；生产：无 query（cookie/JWT） | `SessionContext` | Demo 缺 `memberId` → 400；生产未鉴权 → 401 |

Demo 角色切换时调用 `sessionApi.get(memberId)`；生产环境由 `sessionApi.getCurrent()` 凭 cookie 鉴权。加载失败由 `SessionGate` 展示错误页。

---

### 4.2 Org（组织管理）

客户端：[`org.ts`](../apps/frontend/src/api/org.ts) · Handler：[`org.ts`](../apps/frontend/src/mocks/handlers/org.ts)

#### 数据源 `dataSourceApi`

| 方法 | 路径                            | Body                | 响应                                     |
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

| 方法   | 路径                        | Body / 查询                                                           | 响应                      |
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

### 4.3 Budget（预算管理）

客户端：[`budget.ts`](../apps/frontend/src/api/budget.ts) · Handler：[`budget.ts`](../apps/frontend/src/mocks/handlers/budget.ts)

| 方法   | 路径                                        | Body / 查询                                      | 响应                  |
| ------ | ------------------------------------------- | ------------------------------------------------ | --------------------- | --------------------------------- |
| GET    | `/budget/tree`                              | query: `period?`                                 | `BudgetNode[]`        |
| PUT    | `/budget/nodes/:id`                         | `{ budget, reservedPool? }`                      | `BudgetNode`          |
| GET    | `/budget/departments/:deptId/member-quotas` | —                                                | `MemberBudgetQuota[]` |
| PUT    | `/budget/members/:memberId`                 | `{ personalQuota }`                              | `MemberBudgetQuota`   | 部门内超卖 / 低于已分配 Key → 422 |
| GET    | `/budget/groups`                            | —                                                | `BudgetGroup[]`       |
| POST   | `/budget/groups`                            | `Omit<BudgetGroup, 'id' \| 'consumed'>`          | `BudgetGroup`         |
| PUT    | `/budget/groups/:id`                        | `Partial<Omit<BudgetGroup, 'id' \| 'consumed'>>` | `BudgetGroup`         |
| DELETE | `/budget/groups/:id`                        | —                                                | `void`                |
| GET    | `/budget/overrun-policy`                    | —                                                | `OverrunPolicyConfig` |
| PUT    | `/budget/overrun-policy`                    | `OverrunPolicyConfig`                            | `OverrunPolicyConfig` |
| GET    | `/budget/alerts`                            | —                                                | `AlertRule[]`         |
| POST   | `/budget/alerts`                            | `Omit<AlertRule, 'id'>`                          | `AlertRule`           |
| PUT    | `/budget/alerts/:id`                        | `Partial<AlertRule>`                             | `AlertRule`           |
| DELETE | `/budget/alerts/:id`                        | —                                                | `void`                |

---

### 4.4 Keys（API Key 管理）

客户端：[`keys.ts`](../apps/frontend/src/api/keys.ts) · Handler：[`keys.ts`](../apps/frontend/src/mocks/handlers/keys.ts)

#### 供应商密钥 `providerKeyApi`

| 方法   | 路径                        | Body                      | 响应            |
| ------ | --------------------------- | ------------------------- | --------------- |
| GET    | `/keys/provider`            | —                         | `ProviderKey[]` |
| POST   | `/keys/provider`            | `{ provider, name, key }` | `ProviderKey`   |
| PUT    | `/keys/provider/:id/toggle` | `{ enabled }`             | `void`          |
| POST   | `/keys/provider/:id/rotate` | `{ newKey }`              | `ProviderKey`   |
| DELETE | `/keys/provider/:id`        | —                         | `void`          |

#### 平台密钥 `platformKeyApi`

| 方法   | 路径                           | Body / 查询                                                            | 响应                     | Mock 备注                                                            |
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

| 方法 | 路径                              | Body / 查询                                                   | 响应                                      | Mock 备注                           |
| ---- | --------------------------------- | ------------------------------------------------------------- | ----------------------------------------- | ----------------------------------- |
| GET  | `/keys/approvals`                 | query: `tab?`, `memberId?`                                    | `KeyApproval[]`                           | `tab`: `pending` \| `mine` \| `all` |
| POST | `/keys/approvals`                 | `{ type, reason, requestedQuota, requestedModels, memberId }` | `KeyApproval`                             | 白名单校验 → 422                    |
| PUT  | `/keys/approvals/:id/approve`     | —                                                             | `void`                                    | 预留池不足 → 422                    |
| PUT  | `/keys/approvals/:id/reject`      | `{ reason? }`                                                 | `void`                                    |                                     |
| GET  | `/keys/approvals/:id/quota-check` | —                                                             | `{ sufficient, reservedPool, requested }` |                                     |

---

### 4.5 Models（模型与路由）

客户端：[`models.ts`](../apps/frontend/src/api/models.ts) · Handler：[`models.ts`](../apps/frontend/src/mocks/handlers/models.ts)

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

### 4.6 Dashboard（数据看板）

客户端：[`dashboard.ts`](../apps/frontend/src/api/dashboard.ts) · Handler：[`dashboard.ts`](../apps/frontend/src/mocks/handlers/dashboard.ts)

成本类接口共享查询参数 `CostQueryParams`（`period` 默认 `current_month`）：

| 字段          | 含义                                                              |
| ------------- | ----------------------------------------------------------------- |
| `period`      | `current_month` \| `last_month` \| `last_7_days` \| `custom`      |
| `startDate`   | `period=custom` 时必填，ISO 日期                                  |
| `endDate`     | `period=custom` 时必填，ISO 日期                                  |
| `granularity` | 趋势粒度：`day` \| `week` \| `month`（Mock 返回天级，前端可聚合） |

| 方法 | 路径                                          | 查询                            | 响应                     |
| ---- | --------------------------------------------- | ------------------------------- | ------------------------ |
| GET  | `/dashboard/cost/summary`                     | `CostQueryParams`               | `CostSummary`            |
| GET  | `/dashboard/cost/departments`                 | `parentId?` + `CostQueryParams` | `DepartmentCost[]`       |
| GET  | `/dashboard/cost/departments/:deptId/members` | `CostQueryParams`               | `DepartmentCostMember[]` |
| GET  | `/dashboard/cost/daily`                       | `CostQueryParams`               | `DailyCost[]`            |
| GET  | `/dashboard/cost/top`                         | `limit?` + `CostQueryParams`    | `TopConsumer[]`          |
| GET  | `/dashboard/usage/models`                     | —                               | `ModelUsage[]`           |
| GET  | `/dashboard/usage/teams`                      | —                               | `TeamUsage[]`            |

`parentId` 为空时返回顶层部门成本；指定时返回该部门的子部门成本列表，用于成本钻取。

---

### 4.7 Audit（审计日志）

客户端：[`audit.ts`](../apps/frontend/src/api/audit.ts) · Handler：[`audit.ts`](../apps/frontend/src/mocks/handlers/audit.ts)

| 方法 | 路径                | 查询 / Body                                                                        | 响应                      |
| ---- | ------------------- | ---------------------------------------------------------------------------------- | ------------------------- |
| GET  | `/audit/settings`   | —                                                                                  | `AuditSettings`           |
| PUT  | `/audit/settings`   | `AuditSettings`                                                                    | `AuditSettings`           |
| GET  | `/audit/operations` | `page?`, `pageSize?`, `action?`, `from?`, `to?`, `operatorId?`, `keyword?`         | `Paginated<OperationLog>` |
| GET  | `/audit/calls`      | `page?`, `pageSize?`, `model?`, `status?`, `from?`, `to?`, `callerId?`, `keyword?` | `Paginated<CallLog>`      |

`action` 过滤值见 `AuditAction`；`status` 过滤值：`success` \| `error` \| `filtered`。

---

## 5. 类型参考

完整定义以源码为准。以下列出各域核心类型字段。

### 5.1 Org — [`types/org.ts`](../apps/frontend/src/api/types/org.ts)

**Credential**（ discriminated union by `platform`）：

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

**Member：** `id`, `name`, `phone`, `email`, `departmentId`, `departmentName`, `status`（`active` \| `inactive` \| `pending`）, `roles`, `source`（`imported` \| `manual` \| `invited`）

**BatchImportRow：** `name`, `phone`, `email`, `departmentName`

**MemberBatchImportResult：** `imported`, `failures: { row, reason }[]`

**Role：** `id`, `name`, `type`（`preset` \| `custom`）, `permissions`, `memberCount`

**Permission：** `id`, `name`, `group`

### 5.2 Budget — [`types/budget.ts`](../apps/frontend/src/api/types/budget.ts)

**BudgetNode：** `id`, `name`, `parentId`, `budget`, `consumed`, `reservedPool?`, `children?`, `period`

**MemberBudgetQuota：** `memberId`, `memberName`, `departmentId`, `personalQuota`, `allocated`, `used`

**BudgetGroup：** `id`, `name`, `budget`, `consumed`, `memberIds`, `departmentIds`

**OverrunPolicyConfig：** `thresholds`, `notifyEmail`, `notifyPhone`, `notifyIm`, `blockMessage`

**AlertRule：** `id`, `nodeId`, `nodeName`, `thresholds`, `notifyRoleIds`, `enabled`

**CreateModelInput：** `name`, `displayName`, `baseUrl`, `apiKey`, `inputPrice`, `outputPrice`

**ResolvedWhitelist：** `inherited`, `allowedModels`, `parentCount`

### 5.3 Keys — [`types/keys.ts`](../apps/frontend/src/api/types/keys.ts)

**ProviderType：** `openai` \| `anthropic` \| `deepseek` \| `qwen` \| `custom`

**KeyStatus：** `active` \| `disabled` \| `expired` \| `error`

**ProviderKey：** `id`, `provider`, `name`, `keyPrefix`, `status`, `balance`, `lastUsed`, `createdAt`, `rotateEnabled`

**PlatformKey：** `id`, `name`, `keyPrefix`, `fullKey?`, `memberId`, `memberName`, `appName`, `budgetGroupId`, `budgetGroupName`, `status`, `quota`, `used`, `modelWhitelist`, `createdAt`, `expiresAt`

**ApprovalType：** `key` \| `quota` · **ApprovalStatus：** `pending` \| `approved` \| `rejected`

**KeyApproval：** `id`, `type`, `applicant`, `applicantId`, `department`, `reason`, `requestedQuota`, `requestedModels`, `status`, `approver`, `rejectReason?`, `createdAt`, `resolvedAt`

**MemberQuotaSummary：** `totalQuota`, `used`, `remaining`, `reservedPool`

### 5.4 Models — [`types/models.ts`](../apps/frontend/src/api/types/models.ts)

**ModelInfo：** `id`, `provider`, `name`, `displayName`, `inputPrice`, `outputPrice`, `maxContext`, `enabled`, `capabilities`

**RoutingRule：** `id`, `nodeId`, `nodeName`, `allowedModels`, `defaultModel`, `fallbackModel`, `inherited`

### 5.5 Dashboard — [`types/dashboard.ts`](../apps/frontend/src/api/types/dashboard.ts)

**CostSummary：** `totalCost`, `totalCostMom`, `totalTokens`, `totalRequests`, `totalRequestsMom`, `avgCostPerRequest`, `avgCostPerRequestMom`, `avgCostPerMember`, `avgCostPerMemberMom`

**DepartmentCost：** `departmentId`, `departmentName`, `cost`, `percentage`, `hasChildren?`

**DepartmentCostMember：** `memberId`, `memberName`, `cost`, `requests`, `tokens`

**DailyCost：** `date`, `cost`, `tokens`, `requests`

**TopConsumer：** `memberId`, `memberName`, `department`, `cost`, `tokens`, `requests`

**ModelUsage：** `modelId`, `modelName`, `provider`, `requests`, `tokens`, `cost`, `percentage`

**TeamUsage：** `departmentId`, `departmentName`, `quota`, `consumed`, `memberCount`, `topModel`

### 5.6 Audit — [`types/audit.ts`](../apps/frontend/src/api/types/audit.ts)

**AuditAction：** `key_create` \| `key_disable` \| `key_rotate` \| `budget_change` \| `budget_approve` \| `permission_change` \| `role_assign` \| `model_whitelist_change` \| `member_add` \| `member_remove` \| `org_structure_change`

**OperationLog：** `id`, `action`, `operator`, `operatorId`, `target`, `detail`, `ip`, `createdAt`

**CallLog：** `id`, `caller`, `callerId`, `callerType`（`member` \| `platform_key`）, `model`, `provider`, `inputTokens`, `outputTokens`, `latencyMs`, `status`, `cost`, `createdAt`, `inputPreview`, `outputPreview`

**AuditSettings：** `contentRetentionEnabled`

---

## 6. Mock 与真实 API

### 6.1 开关

| 常量 / 变量             | 定义            | 行为                                         |
| ----------------------- | --------------- | -------------------------------------------- |
| `USE_MOCKS`             | `config/app.ts` | `DEV` 或 `VITE_ENABLE_MOCKS=true` 时启用 MSW |
| `VITE_API_PROXY_TARGET` | 环境变量        | 开发时将 `/api` 代理到真实后端               |

### 6.2 Handler 分工

| 导出              | 用于                       | 行为                                                       |
| ----------------- | -------------------------- | ---------------------------------------------------------- |
| `domainHandlers`  | 域 mock 实现               | session / org / budget / keys / models / dashboard / audit |
| `browserHandlers` | `mocks/browser.ts`         | `domainHandlers` + `fallbackHandlers`                      |
| `serverHandlers`  | `mocks/server.ts` / Vitest | 仅 `domainHandlers`                                        |

- **浏览器：** `fallbackHandlers` 对未匹配的 `/api/*` 返回 `501` JSON，避免 SPA 回退 HTML
- **测试：** `onUnhandledRequest: 'error'`，漏写 handler 立即失败
- **数据：** mock 状态存于 `mocks/data.ts` 内存，刷新页面重置

### 6.3 迁移到真实 API

1. **实现后端** — 按本文档路径与 `api/types/` 实现 REST API
2. **关闭 Mock** — 生产 / 预发设置 `VITE_ENABLE_MOCKS=false`
   - [`.env.production`](../apps/frontend/.env.production)
   - [`vercel.json`](../vercel.json)
   - [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml)
3. **配置网络**
   - 静态托管：将 `/api` 反向代理到后端，或前后端同域部署
   - 本地开发：设置 `VITE_API_PROXY_TARGET=http://localhost:8080` 并设 `VITE_ENABLE_MOCKS=false`
4. **鉴权** — 在 `client.ts` 将 `setDemoMemberIdProvider` 替换为真实 session / token 注入
5. **逐域验收** — session → org → budget → keys → models → dashboard → audit

---

## 7. 变更检查清单

新增或修改 API 时，同步更新：

- [ ] `api/{domain}.ts` — HTTP 方法
- [ ] `api/types/{domain}.ts` — 请求 / 响应类型
- [ ] `api/app-apis.ts` — 若新增域 API 对象
- [ ] `mocks/handlers/{domain}.ts` — mock 实现
- [ ] `mocks/fixtures/` 或 `mocks/data.ts` — 种子 / 可变数据（如需）
- [ ] 本文档 — 端点与类型
- [ ] 页面 Hook / 测试 — 按需补充

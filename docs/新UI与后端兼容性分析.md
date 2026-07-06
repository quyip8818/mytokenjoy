# 新 UI 与后端兼容性分析

> 对比分支：`feature/new-ui-page` vs `main`（后端）  
> 分析日期：2026-07-06

## 结论摘要

| 问题 | 答案 |
|------|------|
| `feature/new-ui-page` 是否修改了后端？ | **否**。仅改动 `apps/frontend/**` 与 `pnpm-lock.yaml` |
| 现有后端能否直接支撑新 UI？ | **部分可以**。核心管理模块（组织、Key、看板、审计等）API 路径大体对齐，但新 UI 分支默认走 MSW Mock，且多处契约与 `main` 后端不一致 |
| 能否不改后端就完整上线新 UI？ | **不能**。预算模型、钱包/成员端、数据源字段映射等需要前端适配或后端补 API |

**总体判断**：现有 Go 后端已覆盖 TokenJoy 管理端的主体能力；新 UI 更像在 Mock 上重新设计的产品原型。合并时应 **保留 `main` 后端不动**，以前端适配为主，仅对明确缺失的 API 做增量开发。

---

## 1. 后端变更情况

### 1.1 `feature/new-ui-page` 相对 `main`

```text
apps/backend/     → 无变更
apps/newapi/      → 无变更
packages/         → 无变更
pnpm-lock.yaml    → 有变更（前端依赖）
apps/frontend/**  → 全部重写
```

该分支从 `0f99fb4 support backend with newapi` 分出，提交 `716eeec feat: new page`（作者 zhouqiao058）。**后端代码与 `main` 完全一致**，不存在需要随 UI 合并而部署的后端变更。

### 1.2 `main` 现有后端 API 一览

后端入口：`apps/backend`，路由注册于 `internal/http/router.go`，业务 API 前缀 `/api`。

| 模块 | 路径前缀 | 主要能力 |
|------|----------|----------|
| 认证 | `/api/auth/*` | 登录、登出、接受邀请 |
| 会话 | `/api/session` | 当前用户、权限、公司上下文 |
| 组织 | `/api/org/*` | 数据源、同步、部门树、成员、角色 |
| 预算 | `/api/budget/*` | 预算树、部门配额、预算组、预警、超限策略 |
| Key | `/api/keys/*` | 供应商 Key、平台 Key、审批 |
| 模型 | `/api/models/*` | 模型列表、启停、路由规则 |
| 看板 | `/api/dashboard/*` | 成本/用量统计、时序 |
| 审计 | `/api/audit/*` | 操作日志、调用日志、设置 |
| 计费 | `/api/billing/*` | 钱包查询、充值 |
| 中继 | `/v1/*` | NewAPI Relay 网关（可选） |
| Webhook | `/api/webhooks/*` | NewAPI 日志入账 |

所有业务接口均依赖 **Session Cookie** + **权限校验**（`CompanyResolve`、`RequireAnyPermission` 等中间件）。

---

## 2. 新 UI 的 API 使用方式

### 2.1 当前分支的实际行为

新 UI 分支在开发模式下 **默认启动 MSW**（`main.tsx` 中 `import.meta.env.DEV` 即启用），请求被 Mock 拦截，**不会打到真实后端**。

同时该分支：

- 删除了登录页与 Session 体系
- `api/client.ts` 无 `credentials: 'include'`，无法携带 Cookie
- `vite.config.ts` 无 `/api` 代理（`main` 通过 `vite-api-proxy` 转发到 `:8080`）

因此，**即使用户本地跑着后端，新 UI 分支开箱即用也不会连上后端**。

### 2.2 新 UI 定义的 API 客户端

新 UI 在 `src/api/*.ts` 中定义了与 Mock handler 一致的 REST 调用。以下按模块对照后端实际能力。

---

## 3. 模块级兼容性对照

图例：

- ✅ 路径与语义基本一致，改前端即可对接
- ⚠️ 路径存在但请求/响应字段有差异，需适配
- ❌ 后端无此 API，需新增或改产品设计
- 🎭 新 UI 页面当前用硬编码 Mock，未调 API

### 3.1 认证与会话

| 新 UI | 现有后端 | 状态 |
|-------|----------|------|
| 无登录页 | `POST /api/auth/login` | ❌ 新 UI 删除认证，无法访问受保护 API |
| 无 Session 处理 | `GET /api/session` | ❌ 所有 `/api/*` 需 Session |

**对接要求**：必须从 `main` 恢复 `features/session/*`、登录页，以及带 Cookie 的 `api/client.ts`。这是接后端的前置条件，与具体页面无关。

---

### 3.2 组织管理 `/org/*`

| API | 新 UI | 后端 | 状态 |
|-----|-------|------|------|
| 数据源状态/测试/保存/导入 | ✅ | ✅ | ✅ |
| 字段映射 GET/PUT/测试 | ✅ | — | ❌ 后端无 `field-mappings` 相关路由 |
| 同步配置/日志/触发 | ✅ | ✅ | ✅ |
| 部门树 CRUD | ✅ | ✅ | ✅ |
| 成员 CRUD/状态/转移/邀请 | ✅ | ✅ | ✅ |
| 角色 CRUD / 成员绑定 | ✅ | ✅ | ✅ |
| 权限列表 | ✅ | ✅ | ✅ |
| 批量邀请/导入 | — | ✅ | 新 UI 未封装，后端已有 |

**说明**：新 UI 数据源向导中的「字段映射」步骤（`step-field-mapping.tsx`）依赖 Mock 的 `/org/data-source/field-mappings`，**后端尚未实现**。组织其余部分可通过适配直接对接；`org/structure` 已用 zustand store 调 `departmentApi` / `memberApi`，是移植成本最低的页面之一。

---

### 3.3 预算管理 `/budget/*`

新 UI 与 `main` 后端在 **预算领域模型上分歧最大**。

| API | 新 UI | 后端 (`main`) | 状态 |
|-----|-------|---------------|------|
| `GET /budget/tree` | ✅ | ✅ | ✅ |
| `PUT /budget/nodes/:id` | ✅ | — | ❌ 后端为 `PUT /budget/departments/{departmentId}` |
| `GET/POST/PUT/DELETE /budget/projects` | ✅ | — | ❌ 后端用 **预算组** ` /budget/groups` 表达类似概念 |
| `GET/PUT /budget/approvals` | ✅ | — | ❌ 后端无预算审批流（仅有 Key 审批） |
| `GET/POST/PUT/DELETE /budget/alerts` | ✅ | ✅ | ⚠️ 字段差异（见下） |

**AlertRule 字段差异**：

| 字段 | 新 UI (`types.ts`) | 后端 (`types/budget.go`) |
|------|-------------------|--------------------------|
| 预警目标 | `targetType` + `targetId` + `targetName` | `nodeId` + `nodeName` |
| 适用范围 | 支持 `team` / `project` | 仅部门节点 |

**后端有、新 UI 未使用的 API**：

- `GET /budget/departments/{id}/member-quotas`
- `PUT /budget/members/{memberId}`（成员配额）
- `GET/POST/PUT/DELETE /budget/groups`
- `GET/PUT /budget/overrun-policy`

**结论**：新 UI 预算页是 **另一套产品模型**（项目预算 + 预算审批），不能零改动对接现有后端。可选路径：

1. **改前端**：新 UI 预算页改用 `budget/groups` + `budget/departments` 等现有 API（推荐，不动后端）
2. **改后端**：新增 `projects`、`approvals`、`nodes` 路由（工作量大，与现有 domain 重复）

---

### 3.4 Key 管理 `/keys/*`

| API | 新 UI | 后端 | 状态 |
|-----|-------|------|------|
| 供应商 Key CRUD/启停/轮换 | ✅ | ✅ | ✅ |
| 平台 Key 列表/创建/吊销/删除 | ✅ | ✅ | ⚠️ 创建参数字段不同 |
| 平台 Key 更新/启停/轮换 | — | ✅ | 新 UI API 层未封装 |
| 审批列表/通过/拒绝 | ✅ | ✅ | ⚠️ 查询参数不同 |
| 审批创建/配额检查 | — | ✅ | 新 UI 未封装 |
| 配额摘要 | — | ✅ | 新 UI 未封装 |

**PlatformKey 创建参数差异**：

| 新 UI | 后端 `CreatePlatformKeyInput` |
|-------|------------------------------|
| `type: 'member' \| 'project'` | 无此字段 |
| `projectId` | 无此字段 |
| `departmentId`（必填） | 无此字段 |
| `quotaMode` | 无此字段 |
| — | `budgetGroupId`（预算组） |
| — | `memberId`（成员 Key） |

后端通过 `memberId` 或 `budgetGroupId` 区分 Key 归属，而非 `type` + `projectId`。新 UI 平台 Key 页需改请求体映射。

**审批列表查询**：

- 新 UI：`?status=pending`
- 后端：`?tab=pending|mine|all` + `memberId`

---

### 3.5 模型路由 `/models/*`

| API | 新 UI | 后端 | 状态 |
|-----|-------|------|------|
| `GET /models` | ✅ | ✅ | ✅ |
| `POST /models` | ✅ | ✅ | ✅ |
| `PUT /models/:id` | ✅ | — | ❌ 后端仅 `PUT /models/{id}/toggle` |
| `DELETE /models/:id` | ✅ | — | ❌ 后端无删除 |
| `GET /models/routing` | ✅ | ✅ | ✅ |
| `PUT /models/routing/:id` | ✅ | ✅ | ✅ |
| `GET /models/routing/resolve` | — | ✅ | 新 UI 未封装 |

新 UI 模型列表页的编辑/删除能力 **超出后端现有接口**，需后端补 API 或前端改为仅支持启停。

---

### 3.6 数据看板 `/dashboard/*`

| API | 新 UI | 后端 | 状态 |
|-----|-------|------|------|
| 成本摘要/部门/日趋势/Top | ✅ | ✅ | ✅ |
| 用量按模型/团队 | ✅ | ✅ | ✅ |
| 用量时序 `usage/series` | — | ✅ | 新 UI 未封装 |
| 部门下钻 `cost/departments/{id}/members` | — | ✅ | 新 UI 未封装 |

看板模块 **兼容性最好**，主要工作是接 `main` 的数据层（React Query）而非改后端。

---

### 3.7 审计日志 `/audit/*`

| API | 新 UI | 后端 | 状态 |
|-----|-------|------|------|
| 操作日志 | ✅ | ✅ | ✅ |
| 调用日志 | ✅ | ✅ | ✅ |
| 审计设置 | — | ✅ | 新 UI 未涉及 |

审计模块可直接对接，注意保留 `main` 的分页/筛选参数约定。

---

### 3.8 钱包 / 计费

| 能力 | 新 UI (`/wallet`) | `main` 前端 (`/billing`) | 后端 |
|------|-------------------|--------------------------|------|
| 路由 | `/wallet` | `/billing` | `/api/billing/wallet` |
| 数据来源 | 🎭 `mockWalletSummary` | `billingApi.getWallet()` | ✅ |
| 充值 | 🎭 Mock UI | `POST /billing/recharge` | ✅ |
| 充值记录/发票/兑换码 | 🎭 Mock | — | ❌ 后端无 |
| 推荐奖励 | 🎭 Mock | — | ❌ 后端无 |

后端 `WalletView` 实际字段：

```json
{
  "balance": 0,
  "allocatable": 0,
  "currency": "CNY",
  "companyId": 1
}
```

新 UI Mock 的 `WalletSummary` 含 `totalConsumed`、`totalRequests`、`invitedCount` 等，**后端不提供**。`main` 前端类型里写的 `availableQuota` 也与后端 `balance` 不一致，合并时需统一字段映射。

**结论**：钱包页可复用 `/api/billing/*` 做余额与充值；充值记录、开票、兑换码等 UI 需等产品定义后再开发后端。

---

### 3.9 成员端 `/me/*`（新 UI 独有）

| 页面 | 路由 | 数据来源 | 后端 |
|------|------|----------|------|
| 成员首页 | `/me` | 🎭 硬编码统计 + 图表 | ❌ 无 member dashboard API |
| 我的 Key | `/me/keys` | 🎭 `mockPlatformKeys` | ⚠️ 可部分复用 |
| 调用日志 | `/me/call-logs` | 未详查，预计 Mock | ⚠️ 可部分复用 |

`main` 已有成员视角入口 `/keys/mine`（同一 `AdminLayout` 下，靠权限区分），后端支持：

- `GET /keys/platform?memberId=...`（当前成员的平台 Key）
- `POST /keys/approvals`（申请 Key）
- `GET /audit/calls?callerId=...`（成员调用日志，需确认权限范围）
- `GET /api/billing/wallet`（成员充值，若开放 `billing:recharge`）

**结论**：成员端是 **纯 UI 原型**，无对应后端模块。短期可复用现有 API 拼装；长期是否需要独立 `MemberLayout` + 专属聚合 API 需产品决策。

---

## 4. 兼容性总览

```text
                    新 UI 页面
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
    可直接对接      需前端适配       需后端新增
    （改接线）      （字段/路径）     （或改设计）
         │               │               │
  组织(除字段映射)   预算(整体)      字段映射 API
  看板             平台 Key 创建    预算 projects
  审计             模型编辑/删除    预算 approvals
  供应商 Key       钱包扩展字段     充值记录/发票
  角色/部门        AlertRule 字段   成员 Dashboard
                   审批查询参数
```

| 等级 | 模块 | 说明 |
|------|------|------|
| 🟢 高 | 组织（结构/角色）、看板、审计、供应商 Key | API 对齐度高，移植 UI 后接 `main` 数据层即可 |
| 🟡 中 | 平台 Key、预警规则、模型（只读+路由）、计费（基础） | 需映射字段或补封装 API |
| 🔴 低 | 预算（项目/审批）、钱包（完整版）、成员端 `/me` | 与后端模型不一致或后端缺失 |

---

## 5. 现有后端能否支撑新 UI？

### 5.1 能支撑的部分（约 60–70% 管理端功能）

在 **恢复 Session/权限/代理** 的前提下，现有后端可支撑新 UI 的：

- 组织架构与成员管理
- 数据源连接、同步（除字段映射步骤）
- 角色与权限
- 供应商 Key 管理
- 平台 Key 列表与基础操作（需适配创建参数）
- Key 审批（需适配查询参数）
- 成本/用量看板
- 操作/调用审计
- 基础钱包余额与充值

这些能力与 `main` 前端已验证过的 API 一致，新 UI 主要是 **换皮 + 交互重做**。

### 5.2 不能直接支撑的部分

| 缺口 | 类型 | 建议 |
|------|------|------|
| 登录/Session | 前端缺失 | 从 `main` 保留，不依赖后端改动 |
| 数据源字段映射 | 后端缺失 | 暂隐藏该步骤，或排期后端实现 |
| 预算项目 + 预算审批 | 模型不一致 | 前端改用 `budget/groups`；或产品确认是否新建后端模块 |
| 模型 PUT/DELETE | 后端缺失 | 前端降级为 toggle；或后端补接口 |
| 钱包充值记录/发票/兑换码 | 后端缺失 | 新 UI 先只做余额+充值，其余隐藏 |
| 成员端 Dashboard | 后端缺失 | 用 dashboard + keys + audit 组合；或新增 BFF 聚合 API |

### 5.3 非后端但阻塞接线的项

即使后端不变，合并新 UI 时也必须从 `main` 带回：

| 项 | 原因 |
|----|------|
| `vite-api-proxy` | 开发环境 `/api` 转发 |
| `api/client.ts`（Cookie + 401/403） | Session 认证 |
| `features/session` + 登录页 | 后端强制认证 |
| `features/query` + `use-apis` | 与现有页面一致的数据层 |
| 关闭 dev 自动 MSW | 否则永远打不到后端 |

---

## 6. 推荐对接策略

### 6.1 原则

1. **后端保持 `main` 不动**，除非产品确认要新增 API
2. **不做整分支 merge**，按模块移植 UI 组件到 `main` 前端
3. **新 UI 的 Mock 仅用于 UI 评审**，接入时删除或隔离到 `VITE_ENABLE_MOCKS`

### 6.2 分阶段计划

| 阶段 | 内容 | 后端改动 |
|------|------|----------|
| P0 | 恢复认证、代理、数据层；移植 Layout/Sidebar | 无 |
| P1 | 组织、看板、审计、供应商 Key | 无 |
| P2 | 平台 Key、Key 审批、模型（只读+toggle+路由） | 可选：模型 update/delete |
| P3 | 预算页改用 `groups` + `departments` API | 无（改前端模型） |
| P4 | 钱包页复用 `/billing`，砍掉 Mock 扩展功能 | 可选：充值记录 |
| P5 | 成员端 `/me` 或复用 `/keys/mine` | 可选：member 聚合 API |

### 6.3 验证清单

```bash
# 后端
cd apps/backend && go run ./cmd/server

# 前端（main 方式，不走 Mock）
cd apps/frontend
VITE_API_PROXY_TARGET=http://127.0.0.1:8080 pnpm dev
```

确认：

1. `POST /api/auth/login` 成功，Cookie 写入
2. `GET /api/session` 返回权限列表
3. 页面 Network 请求指向真实 `:8080`，非 MSW
4. 401 跳转登录页

---

## 7. 附录：后端路由速查

<details>
<summary>组织 <code>/api/org</code></summary>

- `GET  /data-source/status`
- `GET  /data-source/search`
- `POST /data-source/test`
- `PUT  /data-source`
- `POST /data-source/import`
- `POST /data-source/import/retry`
- `GET  /sync/config` · `PUT /sync/config` · `GET /sync/logs` · `POST /sync/trigger`
- `GET  /departments/tree` · `POST /departments` · `PUT /departments/{id}` · `DELETE /departments/{id}`
- `GET  /members` · `POST /members` · `PUT /members/{id}` · `DELETE /members`
- `PUT  /members/status` · `POST /members/transfer` · `POST /members/invite`
- `POST /members/batch-invite` · `POST /members/batch-import`
- `GET  /roles` · `POST /roles` · `PUT /roles/{id}` · `DELETE /roles/{id}`
- `GET  /roles/{roleId}/members` · `POST` · `DELETE /roles/{roleId}/members/{memberId}`
- `GET  /permissions`

</details>

<details>
<summary>预算 <code>/api/budget</code></summary>

- `GET  /tree`
- `PUT  /departments/{departmentId}`
- `GET  /departments/{departmentId}/member-quotas`
- `PUT  /members/{memberId}`
- `GET|POST /groups` · `PUT|DELETE /groups/{id}`
- `GET|PUT /overrun-policy`
- `GET|POST /alerts` · `PUT|DELETE /alerts/{id}`

</details>

<details>
<summary>Key <code>/api/keys</code></summary>

- `GET|POST /provider` · `PUT /provider/{id}/toggle` · `POST /provider/{id}/rotate` · `DELETE /provider/{id}`
- `GET /platform` · `POST /platform` · `PUT /platform/{id}` · `PUT /platform/{id}/toggle`
- `POST /platform/{id}/rotate` · `PUT /platform/{id}/revoke` · `DELETE /platform/{id}`
- `GET /platform/quota-summary`
- `GET /approvals` · `POST /approvals` · `PUT /approvals/{id}/approve|reject`
- `GET /approvals/{id}/quota-check`

</details>

<details>
<summary>其他</summary>

- 模型：`GET /models` · `POST /models` · `PUT /models/{id}/toggle` · `GET /models/routing` · `PUT /models/routing/{id}`
- 看板：`GET /dashboard/cost/*` · `GET /dashboard/usage/*`
- 审计：`GET /audit/operations` · `GET /audit/calls` · `GET /audit/settings`
- 计费：`GET /billing/wallet` · `POST /billing/recharge` · `POST /billing/recharge/{id}/confirm`
- 会话：`GET /session`
- 认证：`POST /auth/login` · `POST /auth/logout` · `POST /auth/accept-invite`

</details>

---

## 8. 相关文档

- [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md)
- [TokenJoy-PRD.md](./TokenJoy-PRD.md)

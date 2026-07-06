# MSW 去除与后端 API 对齐计划

> 分析日期：2026-07-06  
> 修订：同日 v3 — 补充老版 Frontend 运作方式、merge 后现状、重构判定；基础设施改为**从 git 恢复**而非重写  
> 范围：`apps/frontend` MSW、`apps/backend` 已有 REST + seed  
> 老版基线：`be29d28^`（`chore: remove legacy frontend before new UI merge` 的父提交，258 个 `src` 文件）

---

## 1. 结论

| 问题 | 答案 |
|------|------|
| 策略 | **一次性破坏性对齐**：删 MSW、删 MSW 类型与虚构功能，前端字段名 = 后端 JSON 字段名 |
| 要不要重构 frontend？ | **要**（§4）：恢复 merge 前基础设施 + 新 UI 接线；**不**推倒新 UI、**不**重写已对齐的 `budgetApi` |
| 要不要 mapper / compat？ | **不要**。无 `BudgetProjectView`、无 `reserved`↔`reservedPool` 转换层 |
| 要不要 DB migration？ | **不要**。不新增列、不改 `schema.sql`；缺持久化能力则 **删 UI 或延后** |
| 后端要改多少？ | **仅 models PUT/DELETE**（表已有）；其余 **零后端改动** |
| Seed | **改 3 项**：usage 量级、预算组文案、`platform_keys.budgetGroupId` |

**一句话**：把 MSW 层连根拔掉，**恢复老版基础设施**（Session / DI / proxy / `api/types/`），新 UI 页面接入真 API；虚构能力 **删页面**，不为 MSW 补 API 或 migration。

**是否需要重构？** **需要**，但不是推倒重来——见 §4。本质是 **「新 UI 视觉层 + 老版运行时与 API 契约」** 的重组，而非从零写第三套前端。

---

## 2. 老版 Frontend 如何运作（merge 前，`be29d28^`）

> 来源：`git show be29d28^:apps/frontend/...` + [Frontend.md](./Frontend.md)。该版本已接真后端，**无 MSW**。

### 2.1 运行时栈

```text
main.tsx（无 MSW）
  └─ App.tsx
       ├─ /login          lazy LoginPage
       └─ AdminLayout
            ├─ ApiProvider(defaultApis)
            ├─ QueryProvider
            ├─ AuthSessionProvider → SessionGate
            ├─ AuthUnauthorizedBridge（401 → /login）
            ├─ SessionNavigationBridge
            └─ WorkflowProvider → Sidebar / Header / Outlet / WorkflowPanelStack
```

| 能力 | 实现 |
|------|------|
| 本地联调 | `vite-api-proxy.ts` 将 `/api` 反代到 `:8080`；`VITE_API_PROXY_TARGET` 可覆盖 |
| 认证 | `POST /auth/login` → HttpOnly Cookie → `GET /session` |
| 权限 | `GET /session` 一次拉取 permissions；`config/routes.ts` 的 `requiredPermissions` + `PermissionGate` |
| 数据缓存 | TanStack Query + `useInjectedQuery` + `queryKeys` |
| 侧滑表单 | `features/workflow/` Zustand 栈（预算组、Key 审批、成员配额等） |

`pnpm start` 在 monorepo 根目录会同时起 backend + Vite；前端**默认打真 API**，不是 Mock。

### 2.2 分层约定

```text
routes/{domain}/{page}.tsx          薄页面（只拼 JSX，调 Hook）
routes/{domain}/hooks/use-*-page.ts  页面逻辑（useApis / useInjectedQuery）
routes/{domain}/components/         单页展示组件
components/{domain}/                  跨页业务组件
api/{domain}.ts                     HTTP 方法（路径与 backend JSON 一致）
api/types/{domain}.ts               DTO（与 Go domain 1:1，已对齐）
features/session|query|workflow/    横切能力
config/routes.ts                    ROUTE_DEFINITIONS 唯一路由源（17 业务页 + 权限）
```

**页面模式示例**（预算总览）：

```text
overview.tsx
  └─ useBudgetOverviewPage()
       ├─ useBudgetTreeQuery(injectedApis)  → apis.budgetApi.getTree()
       ├─ usePermissions()                  → canAllocate
       └─ useWorkflowRefresh()              → openWithRefresh('budget-node-edit', …)
```

组件层**禁止**直接 `useApis()`（`check-conventions.ts` 会 lint）；数据获取必须在 `use-*-page.ts`。

### 2.3 API 层与依赖注入

```text
页面 Hook
  └─ useInjectedQuery({ queryFn: (apis) => apis.budgetApi.getTree() })
       └─ AppApis（app-apis.ts 聚合 16 个命名空间）
            └─ client.request()  credentials:'include' + 401/403 处理
                 └─ Vite proxy → Go /api/*
```

| 文件 | 职责 |
|------|------|
| `api/client.ts` | `request`、`ApiError`、`buildQuery`、401/403/authz revision 回调 |
| `api/app-apis.ts` | `AppApis` 接口 + `defaultApis` |
| `api/context.tsx` | `ApiProvider` |
| `api/use-apis.ts` | `useApis()` |
| `api/schemas/session.ts` | Zod 校验 session 响应 |
| `api/types/*.ts` | 按域拆分 DTO（如 `BudgetGroup`、`reservedPool`、`nodeId`） |

老版 `budgetApi` **已与后端对齐**（节选）：

- `PUT /budget/departments/{departmentId}`（非 `/budget/nodes`）
- `GET/POST/PUT/DELETE /budget/groups`
- `GET/PUT /budget/overrun-policy`
- `GET .../member-quotas`、`PUT /budget/members/{id}`
- `AlertRule.nodeId` / `nodeName`

钱包走 `billingApi` + 路由 `/billing`（非 `/wallet`）。Key 审批用 `?tab=`。

### 2.4 路由与页面清单（老版 17 页）

| 域 | 路由 | 说明 |
|----|------|------|
| 组织 | `/org/data-source`、`structure`、`roles` | 已接真 API |
| 预算 | `/budget/overview`、`allocation`、`alerts` | 三页拆分；组/部门/配额走 workflow |
| 看板 | `/dashboard/cost`、`usage` | 已接真 API |
| Key | `/keys/provider`、`platform`、`approval`、`mine` | `mine` 成员视角 |
| 模型 | `/models/list`、`routing` | toggle + 路由；无 PUT/DELETE 模型 |
| 计费 | `/billing` | wallet + recharge |
| 审计 | `/audit/operations`、`calls` | 已接真 API |

无 `/me` 成员 Layout、无 MSW、无预算审批 UI。

### 2.5 测试与工程

- Vitest + `createMockApis()` 注入假 API（非 MSW）
- Playwright e2e（`test:e2e`）
- `check-conventions.ts` 约束目录与 `useApis` 用法
- Sentry、`AppErrorBoundary`、路由 lazy + `RouteFallback`

---

## 3. 当前新 UI 现状（`main` HEAD，merge 后）

`9dfe077 merge: adopt new UI frontend` 用 **92 个** `src` 文件替换了老版 **258 个**，净删基础设施约 17k 行。

### 3.1 新 UI 带来了什么

| 项 | 状态 |
|----|------|
| 新视觉 / 组件库 | Radix + 新 layout、预算树交互、成员端 Layout |
| MSW 全量 Mock | `main.tsx` DEV 自动启 worker；开箱不打后端 |
| 新路由形态 | `App.tsx` 硬编码 20+ 路由；无 `config/routes.ts` |
| 新页面 | `/wallet`、`/me/*`、预算单页 `budget/index`（合并 overview+allocation） |
| MSW 领域模型 | `api/types.ts` 单文件：`BudgetProject`、审批、`reserved` 等 |

### 3.2 merge 删掉了什么（需恢复）

| 路径 / 能力 | 影响 |
|-------------|------|
| `vite-api-proxy.ts` | 无法 `/api` 反代 |
| `features/session/*` | 无登录、无 SessionGate、全 API 401 |
| `features/query/*` | 无 TanStack Query 注入模式 |
| `api/app-apis.ts`、`use-apis.ts`、`context.tsx` | 无 DI；测试无法 `createMockApis` |
| `api/auth.ts`、`billing.ts`、`session.ts` | 认证与计费客户端缺失 |
| `api/types/`（按域拆分） | 后端对齐 DTO 被 MSW 单文件 `types.ts` 覆盖 |
| `config/routes.ts`、`nav.ts`、权限门控 | 侧栏硬编码、无按权限隐藏菜单 |
| `features/workflow/*` | 侧滑表单栈消失（预算组、配额、Key 审批等） |
| `routes/*/hooks/use-*-page.ts`（多数） | 页面内联 `useState` + 直调 `budgetApi` / `@/mocks/data` |
| `tests/` 大部分、`playwright`、`check-conventions` | 测试与 lint 约束缩水 |

### 3.3 当前矛盾（同仓库两套范式并存意图）

部分新 UI 组件仍 `import { budgetApi } from '@/api/budget'` 并调用 **MSW 专用**方法（`getProjects`、`updateNode`）；而老版 workflow 文件（若残留）会调 `apis.budgetApi.getGroups`——**API 客户端与调用方契约不一致**。当前 `budget.ts` 仅为 MSW 版，与老版 backend 版不同。

```text
老版（目标恢复）                    当前新 UI（待改）
─────────────────────────────────────────────────────────
api/types/budget.ts                 api/types.ts（MSW 字段）
budgetApi.getGroups()               budgetApi.getProjects()
AdminLayout + 6 层 Provider         AdminLayout 仅 Sidebar+Header
use-*-page.ts + useInjectedQuery    页面内 useEffect + setState
/billing                            /wallet + mockWalletSummary
无 MSW                              DEV 默认 MSW
```

---

## 4. 是否需要重构

### 4.1 结论

| 问题 | 答案 |
|------|------|
| 要不要重构？ | **要**。当前形态无法上线（无 Session、无 proxy、契约错误） |
| 要不要推倒新 UI？ | **不要**。保留新 UI 的 layout / 组件 / 交互 |
| 要不要重写 API 层？ | **不要**。从 `be29d28^` **恢复**老版 `api/` + `features/`，删掉 MSW 字段 |
| 重构性质 | **重组**：基础设施回滚 + 页面接线改造，不是第三套架构 |

### 4.2 三类工作

```text
A. 恢复（git 检出 be29d28^ 对应路径，整目录还原）
   vite-api-proxy、features/session|query|workflow、api/app-apis、api/types/、
   config/routes|nav、auth/login、billing 路由与 hook、check-conventions、tests 骨架

B. 改造（保留新 UI 文件，改调用方式）
   budget/index、wallet→billing、member/*、dashboard、keys、org 页面
   → 拆/写 use-*-page.ts，走 useInjectedQuery，字段对齐 api/types/

C. 删除（MSW 虚构）
   mocks/、BudgetApproval、钱包发票 Tab、field-mappings 网络请求等（同 §6 删除清单）
```

### 4.3 不必做的事

- 不把老版页面 UI 整页换回旧版（新视觉保留）
- 不新建 mapper / view model 层
- 不为 MSW 能力补后端 API 或 DB migration
- 不把 `api/types/` 合并成巨型单文件 `types.ts`

### 4.4 长期架构判定

老版分层（薄页面 + Hook + DI + 后端 1:1 types）**已是长期最优**；新 UI merge 是一次** UI 换代**，误删了运行时。正确终态：

```text
新 UI 组件与路由（视觉）
        +
老版基础设施（Session / proxy / DI / query / workflow / api/types）
        +
删 MSW 虚构域（projects / approvals / mock wallet 扩展）
        =
可上线、可测试、契约不漂移的前端
```

---

## 5. 简化原则（相对 v1）

```text
v1（已废弃）                          v2（本方案）
────────────────────────────────────────────────────────────
mapper / view model                   前端类型与后端 1:1
隐藏 Tab / feature flag               直接删组件与路由
field-mappings API + JSONB 列         向导用静态默认映射，不调 API
可选 status= / 分页包装别名           不做任何后端别名
4 Phase + MSW 可选保留                2 Step，MSW 整包删除
schema 增量 migration                 无 migration；也不改 schema
```

**禁止**：

- Go 别名路由（`/budget/projects`、`/budget/nodes`）
- 为 MSW 虚构能力写 stub handler
- 双轨 MSW `api/types.ts` 与 `api/types/*.ts` 长期并存（恢复后**删除**单文件 `types.ts`）

---

## 6. 删除清单（破坏性，不保留过渡）

### 6.1 整包删除

| 路径 | 说明 |
|------|------|
| `src/mocks/` | `handlers.ts`、`data.ts`、`browser.ts` |
| `public/mockServiceWorker.js` | MSW worker |
| `package.json` 中 `msw` 依赖 | 卸载 |

### 6.2 删类型与 API 方法

从 MSW 单文件 `api/types.ts` **删除**（恢复 `api/types/` 后整文件移除，字段并入各域 types）：

- `BudgetProject`、`BudgetApproval`
- `BudgetNode.reserved` / `memberQuota` / `overrunPolicy`（树节点只保留后端字段）
- `AlertRule.targetType` / `targetId` / `targetName`（改用 `nodeId` / `nodeName`）
- `PlatformKey.type` / `departmentId` / `projectId` / `quotaMode`（改用 `memberId` / `budgetGroupId` / `appName`）
- `CostSummary.monthOverMonth` → `totalCostMom`（及后端其余 `*Mom` 字段）
- `CallLog.inputPreview` / `outputPreview` → `previewSnippet`
- `WalletSummary` / `TopUpRecord` / 发票相关类型
- `Member` 上 HR 扩展字段（`username`、`employeeId` 等）

从 `api/budget.ts` **删除**：

- `getProjects` / `createProject` / `updateProject` / `deleteProject`
- `updateNode`（`/budget/nodes`）
- `getApprovals` / `resolveApproval`

从 `api/keys.ts` **删除或改写**：

- `list` 的 `Paginated` 假设、`departmentId` / `type` 筛选
- `approvalApi.list` 的 `?status=` → `?tab=`

从 `api/org.ts` **删除**：

- `getFieldMappings` / `saveFieldMappings` / `testFieldMapping`（无后端、无 migration）

### 6.3 删 UI 与路由引用

| 组件 / 页面 | 处理 |
|-------------|------|
| `budget-approval-drawer.tsx` | **删除** + 移除所有引用与角标 |
| 钱包页充值记录 / 发票 / 推荐 Tab | **删除** Tab，仅保留余额 + 充值 |
| `budget-edit-allocation` 中 per-node `overrunPolicy` 列 | **删除**列 |
| 预算树 `memberQuota` 列 | **删除**；成员配额只走 `member-quotas` 子页 |
| `@/mocks/data` 全部 import（6 处） | 改 API 或删文件 |

`budget-project-dialog.tsx` → 重命名为 `budget-group-dialog.tsx`，类型改为 `BudgetGroup`。

---

## 7. 前端对齐表（直连，无 mapper）

图例：**删** = 移除功能；**改** = 换路径/字段名；**直连** = 已一致

### 7.1 基础设施（Step 1 最先）

| 项 | 动作 |
|----|------|
| `main.tsx` | 删除 MSW bootstrap；恢复 `initMonitoring()` |
| `vite.config.ts` + `vite-api-proxy.ts` | 从 `be29d28^` **恢复**（勿内联硬编码 proxy） |
| `api/client.ts` | 从 `be29d28^` **恢复**（`credentials`、401/403、JSON 校验） |
| `App.tsx` + `admin-layout.tsx` | 恢复 lazy 路由 + Provider 栈（§2.1） |
| 登录 + `GET /session` | 从 `be29d28^` **恢复** `features/session`、`routes/auth/login`（非重写） |
| `config/routes.ts` | 恢复 `ROUTE_DEFINITIONS`；新 UI 路由（`/me`、`budget/index`）登记入表 |

### 7.2 模块对照

| 模块 | 动作 |
|------|------|
| 认证 / 会话 | **恢复** 老版 session 栈 |
| 组织-数据源 / 同步 / 部门 / 成员 / 角色 | **直连**（恢复 DI 后接 `departmentApi` 等） |
| 组织-字段映射向导 | **改**：静态默认 6 条映射，去掉保存/测试 API 调用；标注「同步引擎接入后再持久化」 |
| 预算-树 | **改**：`BudgetNode` 用 `reservedPool`；删 `memberQuota`、`overrunPolicy` |
| 预算-组 | **改**：`/budget/groups` CRUD，类型 `BudgetGroup`；复用老版 `budget-group-form` workflow |
| 预算-部门更新 | **改**：`PUT /budget/departments/{id}`，`{ budget, reservedPool }` |
| 预算-成员配额 | **改**：`GET .../member-quotas`、`PUT /budget/members/{id}`；复用 `member-quota-config` workflow |
| 预算-超限策略 | **改**：仅 `GET/PUT /budget/overrun-policy`（公司级独立区块） |
| 预算-预警 | **改**：`AlertRule` 直接用 `nodeId`/`nodeName`（`dept-*` 或 `bg-*`） |
| 预算-审批 | **删** |
| 供应商 Key | **直连** |
| 平台 Key | **改**：list 接受裸数组；create body 用 `memberId`/`budgetGroupId`；query 用 `tab` |
| Key 审批 | **改**：`?tab=pending`；复用 `approval-review` workflow |
| 看板 | **改**：`totalCostMom`；`getDailyCosts(?period=current_month)`；`audit` 用 `from`/`to` |
| 审计 calls | **改**：`previewSnippet` |
| 模型 | **直连** list/toggle/routing；编辑走 Step 3 后端 PUT/DELETE |
| 钱包 | **改**：`/wallet` → `/billing`；`billingApi`（`balance`/`allocatable` + recharge） |
| 成员端 `/me` | **改**：keys/logs 接真 API；首页统计无后端则保留静态占位或删图表 |
| 成员端 keys / 日志 | **改**：`keys/platform?memberId=`、`audit/calls?callerId=` |

### 7.3 后端唯一增量

| 路由 | 说明 |
|------|------|
| `PUT /models/{id}` | 更新模型元数据 |
| `DELETE /models/{id}` | 删除模型 |

**不做**：field-mappings、budget/projects、budget/approvals、keys `status=` 别名、platform 分页包装。

---

## 8. 目标 `api/` 结构（恢复老版 + 去 MSW 字段）

> 与 [Frontend.md](./Frontend.md) §4–§5 一致；**不要**收成单文件 `types.ts`。

```text
apps/frontend/src/api/
  client.ts           # 从 be29d28^ 恢复
  app-apis.ts         # AppApis + defaultApis（16 命名空间）
  context.tsx         # ApiProvider
  use-apis.ts
  auth.ts
  session.ts
  org.ts              # 无 field-mappings 方法
  budget.ts           # 从 be29d28^ 恢复（已对齐 backend）
  keys.ts
  dashboard.ts
  audit.ts
  models.ts
  billing.ts
  schemas/session.ts
  types/
    index.ts
    budget.ts         # BudgetGroup、reservedPool、nodeId…
    keys.ts
    org.ts
    dashboard.ts
    audit.ts
    models.ts
    common.ts
```

---

## 9. Seed（3 项）

| 项 | 文件 | 动作 |
|----|------|------|
| 看板量级 | `seed/usage.go` | 扩 `buildUsageBuckets`，`totalCost` 与预算树 consumed 同量级（~67.5k） |
| 预算组文案 | `seed/budget_data.go` | 组名与 UI 文案一致 |
| 平台 Key | `seed/data/platform_keys.json` | 补全 `budgetGroupId: "bg-*"`（当前仅 2/14） |

**不改**：成员数量、`proj-*` ID、预算审批数据、per-node overrunPolicy、MSW 随机曲线。

**usage 分配规则**（可执行）：按 `buildBudgetTree` 各部门 `consumed` 占比，把 ~67.5k 分摊到 `buildUsageBuckets` 各 bucket 的 `CostCNY`；验收 `GET /dashboard/cost/summary` 的 `totalCost` 非 0 且与根节点 consumed 同数量级。

---

## 10. 实施（3 Step）

```text
Step 0 — 恢复基础设施（1–2d）
├── git checkout be29d28^ -- apps/frontend/vite-api-proxy.ts
│   apps/frontend/src/features/ apps/frontend/src/config/
│   apps/frontend/src/api/{app-apis,auth,billing,session,context,use-apis,types,schemas}/
│   apps/frontend/tests/ apps/frontend/scripts/check-conventions.ts
├── 合并冲突：保留新 UI 的 components/、routes/ 视觉文件
├── 恢复 App.tsx 路由壳（login + AdminLayout Provider 栈）
└── 冒烟：pnpm start → login → GET /session → GET /org/departments/tree

Step 1 — 砍 MSW + 删虚构（1d）
├── 删 mocks/、msw 依赖、mockServiceWorker.js、api/types.ts（MSW 单文件）
├── 删 budget-approval、钱包假 Tab、field-mappings API 调用
└── 删 MSW 专用 API 方法（§6.2）

Step 2 — 新 UI 页面接线（3–4d）
├── 预算：budget/index + alerts → use-*-page + workflow + 老版 budgetApi
├── 看板 / keys / org / audit / models：Hook 化 + useInjectedQuery
├── /wallet → /billing；/me/* 接真 API 或占位
├── 后端 models PUT/DELETE + 模型编辑页
├── seed：usage + group 命名 + platform_keys
├── 恢复 check-conventions + vitest/playwright
└── pnpm test 全绿
```

无 Phase 过渡；Step 2 完成即 **MSW 归零**。总工期建议 **5–7 人日**（预算模块单独占 2–3d）。

---

## 11. 验收清单

### 11.1 基础设施（老版能力回归）

- [ ] `AdminLayout` 含 ApiProvider + QueryProvider + AuthSessionProvider + WorkflowProvider
- [ ] `vite-api-proxy.ts` 生效；无 MSW bootstrap
- [ ] `useApis()` / `createMockApis()` 可用于测试
- [ ] `config/routes.ts` 驱动侧栏与权限

### 11.2 无 MSW

- [ ] 无 `mocks/`、无 `msw` 依赖、无 `VITE_ENABLE_MOCKS`
- [ ] 无 `from '@/mocks/data'`
- [ ] 无 `api/types.ts`（MSW 单文件）；契约在 `api/types/*.ts`

### 11.3 真 API

- [ ] `GET /api/budget/groups` 驱动预算组列表
- [ ] `PUT /api/budget/departments/dept-3` 更新 `reservedPool`
- [ ] `GET/PUT /api/budget/overrun-policy`
- [ ] `POST /api/budget/alerts`  body 含 `nodeId`
- [ ] `GET /api/keys/approvals?tab=pending`
- [ ] 看板 `totalCost` 非 0
- [ ] `PUT/DELETE /api/models/{id}`

### 11.4 已删除（非隐藏）

- [ ] 无预算审批入口与组件
- [ ] 无钱包充值记录 / 发票 Tab
- [ ] 无 field-mappings 网络请求
- [ ] 预算树无 per-node overrun / memberQuota 可编辑列

```bash
pnpm start
pnpm -F @tokenjoy/backend test
pnpm -F @tokenjoy/frontend test
```

---

## 12. 风险与取舍

| 取舍 | 说明 |
|------|------|
| git 恢复 vs 手写 | 恢复老版可降低 Session/DI 回归风险；合并时以新 UI 视觉文件为准 |
| 字段映射不持久化 | 无 migration 的代价；向导可演示，同步引擎上线前再补 API + 列 |
| 预算 UI 文案仍写「项目」 | 仅文案，代码类型统一 `BudgetGroup` |
| `/me` 首页无聚合 API | 短期静态占位；长期可复用 `/keys/mine` 或新增 BFF |
| 破坏性 types 改动 | 删 MSW `types.ts`，统一 `api/types/`；不做 alias 过渡 |
| 登录链路缺失 | Step 0 必做，否则真 API 全 401 |
| 契约长期漂移 | 恢复后补 schema 测试或 OpenAPI（Roadmap，非本次阻塞） |

---

## 13. 相关文档

- [新UI与后端兼容性分析.md](./新UI与后端兼容性分析.md)
- [Frontend.md](./Frontend.md)
- [NewAPI-集成状态与缺口.md](./NewAPI-集成状态与缺口.md)

---

## 14. 一句话总结

**恢复 merge 前老版运行时（Session / proxy / DI / `api/types/` / workflow），删 MSW 与虚构 domain，新 UI 页面按 Hook + 真 API 接线；不做 migration、不做 mapper、不做后端别名；后端只补 models PUT/DELETE；seed 调 usage 量级与组/Key 关联。**

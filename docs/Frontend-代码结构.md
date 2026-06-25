# TokenJoy 前端架构

`apps/frontend` 的架构说明与开发规范。描述当前代码组织、运行时行为与扩展方式，供日常开发与 Code Review 参考。

---

## 0. 相关文档

| 文档 | 路径 | 职责 |
| --- | --- | --- |
| API 契约 | [Frontend-API契约.md](./Frontend-API契约.md) | REST 路径、请求/响应体、分页、错误格式、Mock 切换 |
| Demo 交互设计 | [Demo-交互设计方案.md](./Demo-交互设计方案.md) | Workflow 侧滑、Demo 引导、CTA 高亮 |
| 开发速查 | [CLAUDE.md](./CLAUDE.md) | 命令、技术栈、目录一览 |
| 产品需求 | [TokenJoy-PRD.md](./TokenJoy-PRD.md) | 业务域边界、功能范围 |
| Cursor 规范 | [`.cursor/rules/frontend-structure.mdc`](../.cursor/rules/frontend-structure.mdc) | AI / 新人速查摘要 |

**边界：** 类型以 `api/types/` 为准；Workflow 交互见 Demo 设计文档；Mock 路径须与 API 契约同步。

---

## 1. 技术栈

| 类别 | 选型 |
| --- | --- |
| 框架 | React 19、React Router 7 |
| 构建 | Vite 8、TypeScript 6 |
| 样式 | Tailwind CSS 4、shadcn/ui（`components/ui/`） |
| 表格 | TanStack Table |
| 图表 | Recharts（数据转换在页面 Hook） |
| 表单 | react-hook-form |
| 状态 | Zustand（workflow、demo、page subtitle） |
| Mock | MSW 2（浏览器 Service Worker + Node `setupServer`） |
| 测试 | Vitest、Testing Library |

路径别名：`@/` → `src/`，`@tests/` → `tests/`。

---

## 2. 运行时架构

### 2.1 启动流程

```
main.tsx
  ├─ USE_MOCKS ? startBrowserMockWorker() : skip
  │    └─ 失败 → 全屏错误页（提示清除 Service Worker）
  └─ createRoot → App.tsx
```

`USE_MOCKS` 定义于 `config/app.ts`：开发环境默认开启，生产构建通过 `VITE_ENABLE_MOCKS=true`（`build:gh-pages`）开启 Demo Mock。

MSW 浏览器 worker 经 `mocks/start-browser-worker.ts` 启动，Service Worker 路径与 scope 使用 `SERVICE_WORKER_URL`、`SERVICE_WORKER_SCOPE`（均基于 `import.meta.env.BASE_URL`），以支持 GitHub Pages 子路径部署。

### 2.2 Provider 树

```
App (BrowserRouter + lazy routes)
└─ AdminLayout
   ├─ ApiProvider          defaultApis
   ├─ DemoProvider         角色切换 + 引导 + 导航桥接
   ├─ WorkflowProvider     侧滑栈状态（Zustand）
   ├─ Sidebar / Header / Outlet
   ├─ WorkflowPanelStack   全局侧滑面板
   └─ Toaster
```

首页 `/` 由 `HomeRedirect` 根据当前 Demo 角色权限跳转到 `HOME_PATH_CANDIDATES` 中第一个可访问页面。

### 2.3 环境与部署

| 配置 | 来源 | 作用 |
| --- | --- | --- |
| `BASE_URL` | Vite `base` | 路由 basename、静态资源、API 前缀 |
| `API_BASE_PATH` | `{BASE_URL}/api` | `api/client.ts` 请求前缀 |
| `VITE_API_PROXY_TARGET` | 环境变量 | 开发时 `/api` 代理到真实后端 |
| `GITHUB_REPOSITORY` | CI / 本地 preview | 推导 GitHub Pages base（`/{repo}/`） |

`vite.config.ts` 在 GitHub Pages 构建时复制 `index.html` → `404.html`，支持 SPA 回退。`deploy.yml` 使用 `pnpm build:gh-pages` 部署带 Mock 的 Demo 包。

---

## 3. 目录结构

```
apps/frontend/
├── public/                 mockServiceWorker.js（MSW 生成）
├── scripts/
│   └── check-conventions.mjs   路由 / 页面 Hook / ui 域名校验
├── tests/                  镜像 src 子路径的 Vitest 用例
└── src/
    ├── main.tsx            启动入口
    ├── App.tsx             路由注册（lazy + Suspense）
    ├── config/
    │   ├── routes.ts       ROUTES / ROUTE_META / APP_ROUTES
    │   ├── nav.ts          NAV_GROUP_LAYOUT → NAV_GROUPS
    │   └── app.ts          环境常量
    ├── routes/{domain}/    页面入口 + hooks/ + components/
    ├── components/
    │   ├── ui/             无业务语义的 primitive
    │   ├── layout/         壳层、导航、数据区
    │   ├── auth/           权限门控
    │   └── {domain}/       跨页或 workflow 复用的业务组件
    ├── features/
    │   ├── workflow/       侧滑编排（Zustand + 面板栈）
    │   └── demo/           Demo 角色、引导、Chrome
    ├── api/                HTTP 客户端 + 聚合注入
    ├── hooks/              跨域通用 React Hook
    ├── lib/                纯函数、常量、权限工具
    └── mocks/              MSW handlers、fixtures、启动辅助
```

**分层原则：**

- `routes/` — 与 `APP_ROUTES` 一一对应的页面域
- `components/{domain}/` — 多页面或 workflow 共享的业务 UI
- `features/` — 跨域、独立 Provider、或可整体开关的能力（workflow、demo）
- `api/` + `lib/` — 无 UI 的业务与基础设施

---

## 4. 路由体系

[`config/routes.ts`](../apps/frontend/src/config/routes.ts) 为唯一路由元数据源。

| 导出 | 职责 |
| --- | --- |
| `ROUTES` | 路径常量（业务代码禁止硬编码 `/org/...` 等字符串） |
| `ROUTE_META` | path、label、icon、`requiredPermissions`、可选 `badgeKey` |
| `APP_ROUTES` | lazy 页面模块，供 `App.tsx` 注册 |
| `HOME_PATH_CANDIDATES` | 首页跳转优先级 |
| `toRouterPath` | 去掉 leading `/` 供 React Router `path` 使用 |

[`config/nav.ts`](../apps/frontend/src/config/nav.ts) 的 `NAV_GROUP_LAYOUT` 只负责侧栏分组与顺序，条目元数据从 `ROUTE_META` 派生。

**新增页面清单：** `ROUTES` → `ROUTE_META` → `APP_ROUTES` → `NAV_GROUP_LAYOUT` → `{page}.tsx` + `hooks/use-{page}-page.ts`。`pnpm lint` 中的 `check-conventions` 会校验四者对齐及页面 Hook 存在性。

当前共 **16** 个业务页面（不含首页重定向）：

| 域 | 页面 | Hook |
| --- | --- | --- |
| dashboard | cost、usage | `use-cost-dashboard-page`、`use-usage-dashboard-page` |
| org | data-source、structure、roles | `use-data-source-page`、`use-structure-page`、`use-roles-page` |
| budget | overview、allocation、alerts | `use-budget-overview-page`、`use-budget-allocation-page`、`use-budget-alerts-page` |
| keys | mine、approval、platform、provider | `use-my-keys-page`、`use-approval-page`、`use-platform-keys-page`、`use-provider-keys-page` |
| models | list、routing | `use-model-list-page`、`use-model-routing-page` |
| audit | operations、calls | `use-audit-operations-page`、`use-audit-calls-page` |

keys 域另有共享 Hook `use-keys-list-page`，供 platform / provider 列表页复用。

---

## 5. API 层

### 5.1 结构

```
api/
├── client.ts           request()、ApiError、buildQuery()
├── app-apis.ts         AppApis 接口 + defaultApis 聚合
├── api-context.ts      React Context
├── context.tsx         ApiProvider
├── use-apis.ts         useApis()
├── {domain}.ts         各资源 HTTP 方法
└── types/{domain}.ts   DTO / 响应类型
```

所有请求经 `request()` 发往 `API_BASE_PATH`。Demo 模式下 `setDemoMemberIdProvider` 注入 `X-Demo-Member-Id` 请求头，与 `sessionApi` 角色切换联动。

### 5.2 依赖注入

- 生产：`AdminLayout` 注入 `defaultApis`
- 页面 Hook：`const apis = injectedApis ?? useApis()`，测试传入 mock
- 非 React：`createDemoRoleStore(..., { sessionApi })` 等同理注入

**改 API 时同步：** `api/{domain}.ts`、`api/types/`、`mocks/handlers/`、fixtures（如需）、[Frontend-API契约.md](./Frontend-API契约.md)。

---

## 6. 页面架构

标准三层：**薄页面 → 页面 Hook → 展示组件**。

### 6.1 页面（`{page}.tsx`）

只负责组合布局与展示组件，从 Hook 取 view model：

```tsx
export default function ExamplePage() {
  const vm = useExamplePage()
  return (
    <PageShell>
      <DataSection loading={vm.loading} error={vm.error} onRetry={vm.refresh}>
        <ExampleTable {...vm.table} />
      </DataSection>
    </PageShell>
  )
}
```

### 6.2 页面 Hook（`hooks/use-{page}-page.ts`）

- 调用 `useApis()`（支持 `injectedApis`）
- 用 `useAsyncResource` / `useFilteredResource` 管理异步数据
- 编排 workflow 打开、筛选、分页、行选等
- 返回扁平 view model，**不返回 JSX**

Workflow 关闭后刷新列表：`useWorkflowRefresh(refresh)`。

### 6.3 展示组件

- props 受控，不直接 `import { xxxApi }`
- 单页专用 → `routes/{domain}/components/`
- 跨页 / workflow → `components/{domain}/`

---

## 7. features

### 7.1 Workflow（`features/workflow/`）

侧滑多步表单的编排层，基于 Zustand `workflow-store`。

| 模块 | 职责 |
| --- | --- |
| `workflow-definitions.tsx` | 所有 workflow 注册表（id → 组件 + 默认 layer + title） |
| `workflow-payloads.ts` | 各 workflow 的 open 参数类型 |
| `workflows/*.tsx` | 具体步骤面板（可复用 `components/{domain}/` 表单） |
| `components/workflow-panel-stack.tsx` | 全局侧滑栈渲染 |
| `use-workflow.ts` | `open` / `push` / `pop` / `closeAll` |
| `use-workflow-submit.ts` | 提交流程封装 |

新增 workflow：`workflows/{name}.tsx` + `workflow-payloads.ts` 类型 + `workflow-definitions.tsx` 注册。

### 7.2 Demo（`features/demo/`）

Demo 可整体通过 `USE_MOCKS` 与 `DemoProvider` 开关。

| 子模块 | 职责 |
| --- | --- |
| `roles/` | 角色切换 store、`useDemoRole`、`HomeRedirect`、导航桥接 |
| `guide/` | 分步引导、CTA 高亮、`DemoGuidePanel` |
| `chrome/` | Demo 横幅、工具栏、桌面端提示 |
| `nav/use-approval-pending-count` | 审批待办角标 |

`usePermissions()` 读取 demo session 权限，驱动侧栏可见性与 `PermissionGate`。

---

## 8. 共享组件索引

### 8.1 布局与通用

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `AdminLayout`、`Sidebar`、`Header` | `components/layout/` | 应用壳 |
| `PageShell`、`DataSection` | `components/layout/` | 页面容器与异步态 |
| `RouteFallback` | `components/layout/` | lazy 路由加载态 |
| `PermissionGate` | `components/auth/` | 权限门控 |
| `EmptyState`、`ErrorState`、`ConfirmActionDialog` | `components/ui/` | 通用反馈 |
| `OptionsSelect` | `components/ui/` | 无业务语义的枚举筛选下拉 |

### 8.2 跨域业务组件

| 组件 | 路径 | 消费者 |
| --- | --- | --- |
| `CredentialForm`、`SyncConfigPanel` | `components/org/` | workflow |
| `BudgetProgressCell` | `components/budget/` | budget、usage、platform keys |
| `AuditFilteredPage`、`AuditToolbar` | `components/audit/` | operations、calls |
| `AuditMemberSelect`、`AuditDatePresetSelect`、`AuditKeywordInput` | `components/audit/` | operations、calls |
| `OptionsSelect` | `components/ui/` | audit 筛选、可复用于 dashboard |

### 8.3 页面私有组件（按域）

| 域 | 组件 | 页面 |
| --- | --- | --- |
| org | `structure-*`、`department-tree`、`member-table`、`role-*`、`data-source-*`、`import-result`、`sync-log-table` | structure、roles、data-source |
| keys | `my-keys-table`、`approval-table`、`platform-key-table`、`provider-key-table` | mine、approval、platform、provider |
| budget | `budget-row`、`budget-group-table` | overview、allocation |
| dashboard | `cost-*` ×5、`usage-model-chart` | cost、usage |
| models | `model-list-table`、`routing-rules-table` | list、routing |
| audit | `call-logs-table`、`operations-log-table` | calls、operations |

---

## 9. hooks 与 lib

### 9.1 全局 Hook（`hooks/`）

| Hook | 职责 |
| --- | --- |
| `useAsyncResource` | loading / error / refresh / `setData` |
| `useFilteredResource` | 带 filter 的列表资源 |
| `useWorkflowRefresh` | workflow 成功后触发 refresh |
| `usePermissions` | 权限与 `canWrite` |
| `useRowHighlight` | 表格行高亮 |
| `usePageSubtitle` | Header 副标题（Zustand） |
| `useAuditSettings` | 审计页共享筛选设置 |
| `useAuditMemberOptions` | 审计成员筛选下拉选项 |

Audit 列表页共享 Hook：`routes/audit/hooks/use-audit-list-page.ts`（内部 `useFilteredResource`）。

### 9.2 纯逻辑（`lib/`）

| 模块 | 内容 |
| --- | --- |
| `permission-keys.ts` | 权限 key 常量 |
| `permissions.ts` | `hasPermission`、`getDefaultHomePath` 等 |
| `labels.ts`、`label-badges.tsx` | 业务标签与徽章映射 |
| `demo-clock.ts` | Demo 时间锚点（`DEMO_TODAY`、`resolveLast7DaysRange`） |
| `audit-constants.ts`、`audit-query.ts` | 审计日期 preset 与 query 构建 |
| `{domain}.ts` | 域内纯函数（org、budget、dashboard） |
| `utils.ts` | `cn()` 等工具 |
| `csv-export.ts` | 导出辅助 |

常量与环境：`config/routes.ts`、`config/app.ts`、`lib/demo-clock.ts`、`features/workflow/constants.ts`、`features/demo/guide/constants.ts`。

---

## 10. Demo API 与测试

Demo 模式下，`/api/*` 由 MSW 在浏览器内实现（**Demo API**），返回内存 fake 数据；`api/*.ts` 与生产共用同一调用链，不访问外部后端。

**Handler 分工**（[`handlers/index.ts`](../apps/frontend/src/mocks/handlers/index.ts)）：

| 导出 | 用于 | 行为 |
| --- | --- | --- |
| `domainHandlers` | 域 mock 实现 | session / org / budget / … |
| `browserHandlers` | [`browser.ts`](../apps/frontend/src/mocks/browser.ts) | `domainHandlers` + `fallbackHandlers`（未匹配 API → 501 JSON，避免 SPA 回退 HTML） |
| `serverHandlers` | [`server.ts`](../apps/frontend/src/mocks/server.ts) / Vitest | 仅 `domainHandlers`；未 mock 请求触发 `onUnhandledRequest: 'error'` |

Demo session 加载失败时通过 `sessionError` + `DemoSessionGate` 展示 `ErrorState`，不静默清空为无权限空会话。

### 10.1 MSW 布局

```
mocks/
├── browser.ts              setupWorker（浏览器）
├── server.ts               setupServer（Vitest）
├── start-browser-worker.ts 浏览器启动封装
├── handlers/{domain}.ts    与 api 同域
├── handlers/fallback.ts    未匹配 API 路径 → 501 JSON
├── fixtures/{domain}.ts    静态种子数据
├── data.ts                 可变内存 store
└── lib/                    paginate、query、parse 等 mock 工具
```

- 开发 / GitHub Pages：`startBrowserMockWorker()` 使用 `browserHandlers`，`onUnhandledRequest: 'bypass'`（静态资源 bypass；API 由 domain + fallback 消化）
- 测试：`tests/setup.ts` 挂载 `serverHandlers`，`onUnhandledRequest: 'error'`（漏写 handler 立即失败）

Mock 工具仅放 `mocks/lib/`，禁止进入 `src/lib/`。

### 10.2 测试布局

```
tests/
├── setup.ts、utils.tsx       MSW + renderHookWithProviders / createMockApis
├── fixtures/、helpers/
├── lib/、hooks/、config/
├── routes/{domain}/
└── components/{domain}/
```

增量测试优先级：`lib/` 纯函数 → 全局 Hook → 页面 Hook（`injectedApis`）→ 关键组件。

---

## 11. 代码放置决策

| 我要写的代码 | 放哪里 | 判断条件 |
| --- | --- | --- |
| 路由页面入口 | `routes/{domain}/{page}.tsx` | 每个 `APP_ROUTES` 一条 |
| 页面逻辑 | `routes/{domain}/hooks/use-{page}-page.ts` | 状态、副作用、编排 |
| 单页 UI 块 | `routes/{domain}/components/` | 仅 1 个 route，且无 workflow 复用 |
| 跨页 / workflow UI | `components/{domain}/` | ≥2 页面，或页面 + workflow |
| 布局 / 权限 | `components/layout/`、`auth/` | 跨业务域 |
| Primitive | `components/ui/` | 无业务语义 |
| Workflow 面板 | `features/workflow/workflows/` | `useWorkflow().open()` |
| HTTP | `api/{domain}.ts` | 对应后端资源 |
| DTO | `api/types/{domain}.ts` | 与 api 同域 |
| 纯逻辑 | `lib/` | 无 React，可单测 |

**升降级：** 第二页面引用 `routes/*/components/` → 升到 `components/{domain}/`；长期单页且无 workflow → 降到 `routes/*/components/`。

---

## 12. 工具链

| 命令 | 作用 |
| --- | --- |
| `pnpm start` | 本地开发（Mock 默认开） |
| `pnpm lint` | ESLint + `check-conventions` |
| `pnpm test` / `pnpm test:run` | Vitest（含 typecheck:test） |
| `pnpm build` | `tsc -b && vite build` |
| `pnpm build:gh-pages` | `VITE_ENABLE_MOCKS=true` 生产构建 |
| `pnpm preview:gh-pages` | 模拟 GitHub Pages 子路径预览 |

**CI：** [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) 执行 lint、test、带 Mock 的 build；[`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) 部署 GitHub Pages Demo。

---

## 13. 反模式（Review 对照）

| 反模式 | 正确做法 |
| --- | --- |
| 页面内大量 `useState` + handler | `use-*-page.ts` |
| 子组件直接 `import { xxxApi }` | `useApis()` 或 Hook 传回调 |
| `components/ui` 含业务语义 | `components/{domain}/` 或 `routes/.../components/` |
| 单页组件放 `components/{domain}/` | `routes/{domain}/components/` |
| 硬编码路由字符串 | `ROUTES.*` |
| 新路由未同步 config | 更新 `ROUTE_META` / `APP_ROUTES` / `NAV_GROUP_LAYOUT` |
| 忽略 `useAsyncResource` 的 `error` | `ErrorState` 或 toast |
| Mock 与 api 路径不一致 | 对照 API 契约 |
| mock 工具进 `src/lib/` | `mocks/lib/` |
| 映射表多处定义 | `lib/labels.ts` 或 domain constants |
| 筛选下拉硬编码 label / value | `lib/labels.ts` + `components/ui/options-select.tsx` |
| Audit 列表手写 filter state | `use-audit-list-page` + `useFilteredResource` |
| Demo 日期多处硬编码 | `lib/demo-clock.ts` |

---

## 14. 扩展清单

**新功能：** Page → Hook → Components；路由走 §4 清单；组件按 §8 与 §11 选型。

**新 API：** 契约 + handler + 类型同步；页面 Hook 经 `useApis()` 消费；按需补测试。

**新 Workflow：** `workflows/{name}.tsx` + payload 类型 + `workflow-definitions.tsx` 注册；可复用 UI 放 `components/{domain}/`。

**提 PR 前：** `pnpm lint` 与 `pnpm test` 本地通过。

# TokenJoy 前端代码结构优化指南

`apps/frontend` 的日常开发与 Code Review 参考手册：规定代码放哪里、怎么拆、有哪些门禁。历史重构（页面 Hook、ApiProvider、ROUTE_META、测试 CI、组件边界）已全部落地，下文仅保留**持续生效**的规范与索引。

---

## 0. 相关文档

| 文档          | 路径                                                                              | 职责                                              |
| ------------- | --------------------------------------------------------------------------------- | ------------------------------------------------- |
| API 契约      | [Frontend-API契约.md](./Frontend-API契约.md)                                      | REST 路径、请求/响应体、分页、错误格式、Mock 切换 |
| Demo 交互设计 | [Demo-交互设计方案.md](./Demo-交互设计方案.md)                                    | Workflow 侧滑、Demo 引导、CTA 高亮                |
| 开发速查      | [CLAUDE.md](./CLAUDE.md)                                                          | 命令、技术栈、目录一览                            |
| 产品需求      | [TokenJoy-PRD.md](./TokenJoy-PRD.md)                                              | 业务域边界、功能范围                              |
| Cursor 规范   | [`.cursor/rules/frontend-structure.mdc`](../.cursor/rules/frontend-structure.mdc) | AI / 新人速查摘要                                 |

**边界说明：** 类型以 `api/types/` 为准；Workflow 交互见 Demo 设计文档；Mock 路径须与 API 契约同步。

---

## 1. 架构概览

### 1.1 目录分层

```
apps/frontend/src/
├── config/           # routes.ts（ROUTES / ROUTE_META / APP_ROUTES）、nav.ts、app.ts
├── routes/{domain}/  # {page}.tsx + hooks/ + components/
├── components/       # ui/、layout/、auth/、{domain}/（跨页或 workflow 复用）
├── features/         # workflow、demo
├── api/              # HTTP + types/（useApis 消费）
├── hooks/            # 全局通用 Hook
├── lib/              # 纯函数、常量、权限
└── mocks/            # MSW handlers + fixtures
```

技术栈：React 19、React Router 7、Vite、Tailwind 4、Zustand（workflow/demo）、MSW、Vitest。

### 1.2 Provider 树

`main.tsx`（可选 MSW）→ `App.tsx` → `AdminLayout` → `ApiProvider` → `DemoProvider` → `WorkflowProvider` → `Outlet` + 侧滑栈 + Toaster。

Mock 开关：`config/app.ts` 中 `USE_MOCKS`；API 前缀：`API_BASE_PATH`。

### 1.3 核心模式（摘要）

| 模式               | 位置                                                           |
| ------------------ | -------------------------------------------------------------- |
| 薄页面 + 页面 Hook | `routes/*/{page}.tsx` + `hooks/use-*-page.ts`                  |
| 路由元数据单一来源 | `ROUTE_META`（path / label / icon / 权限）+ `NAV_GROUP_LAYOUT` |
| API 依赖注入       | `ApiProvider` + `useApis()`                                    |
| 异步数据           | `useAsyncResource`；列表筛选 `useFilteredResource`             |
| 侧滑编排           | `features/workflow/` + `useWorkflowRefresh`                    |

---

## 2. 代码放置决策表

| 我要写的代码       | 放哪里                                     | 判断条件                               |
| ------------------ | ------------------------------------------ | -------------------------------------- |
| 路由页面入口       | `routes/{domain}/{page}.tsx`               | 每个 `APP_ROUTES` 一条                 |
| 页面逻辑           | `routes/{domain}/hooks/use-{page}-page.ts` | 状态、副作用、编排                     |
| 单页 UI 块         | `routes/{domain}/components/`              | 仅 1 个 route 页面，且无 workflow 复用 |
| 跨页 / workflow UI | `components/{domain}/`                     | ≥2 页面，或 1 页面 + workflow          |
| 布局 / 权限门控    | `components/layout/`、`auth/`              | 跨业务域                               |
| Primitive          | `components/ui/`                           | 无业务语义                             |
| Workflow 面板      | `features/workflow/workflows/`             | `useWorkflow().open()`                 |
| HTTP               | `api/{domain}.ts`                          | 对应后端资源                           |
| DTO                | `api/types/{domain}.ts`                    | 与 api 同域                            |
| 纯逻辑             | `lib/`                                     | 无 React，可单测                       |

**组件升级/降级：** 第二页面引用 `routes/*/components/` → 升到 `components/{domain}/`；`components/{domain}/` 长期单页且无 workflow → 降到 `routes/*/components/`。

**features/ 准入：** 跨域、独立 Provider、或可整体开关（workflow、demo）；单域页面 UI 留 `routes/` + `components/`。

---

## 3. 页面开发模板

**`{page}.tsx`** — 只组合：

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

**`hooks/use-example-page.ts`** — `useApis()`（可 `injectedApis` 测）、`useAsyncResource`、返回扁平 view model，不返回 JSX。

**展示组件** — props 受控；不直接 `import { xxxApi }`（workflow 内可复用 `components/{domain}/` 表单）。

---

## 4. 路由与配置

[`config/routes.ts`](../apps/frontend/src/config/routes.ts) 为唯一路由元数据源：

1. 路径只在 `ROUTES`；业务代码用 `ROUTES.*`（ESLint 禁止硬编码 `/org|budget|...` 路径）
2. `ROUTE_META`：path、label、icon、permissions、可选 `badgeKey`
3. `NAV_GROUP_LAYOUT`（[`nav.ts`](../apps/frontend/src/config/nav.ts)）：仅分组与顺序
4. 新页面同步：`ROUTES` → `ROUTE_META` → `APP_ROUTES` → `NAV_GROUP_LAYOUT` → 页面 + `use-*-page.ts`
5. `pnpm lint` 含 `check-conventions`（路由对齐 + 页面 Hook + `ui/` 域名检查）

---

## 5. API 与依赖注入

- 聚合：[`api/app-apis.ts`](../apps/frontend/src/api/app-apis.ts) → `AppApis` / `defaultApis`
- 消费：`const apis = useApis()`；测试注入 `injectedApis ?? useApis()`
- 非 React：`createDemoRoleStore(..., { sessionApi })`

**改 API 时同步：** `api/{domain}.ts`、`api/types/`、`mocks/handlers/`、fixtures（如需）、[Frontend-API契约.md](./Frontend-API契约.md)。

---

## 6. 数据与错误处理

- 页面数据：`useAsyncResource` 放在页面 Hook；`deps` 声明依赖
- Workflow 关闭刷新：`useWorkflowRefresh(refresh)`
- 带筛选列表：`useFilteredResource`
- 展示错误：`DataSection` / `ErrorState` + `onRetry`；操作类用 `sonner` toast
- 除 workflow 子面板外，避免组件内 `useEffect` 拉数

---

## 7. 组件规范

### 7.1 何时拆分

| 信号                   | 动作                                                                    |
| ---------------------- | ----------------------------------------------------------------------- |
| 页面 > 120 行          | 抽 `use-*-page.ts`                                                      |
| 内联 Row/Chart/Toolbar | 单页 → `routes/.../components/`；多页/workflow → `components/{domain}/` |
| 单文件 > 200 行        | 必须拆分                                                                |

### 7.2 API 设计

受控 props；表格 `readOnly` / `rowSelection` 由 Hook 传入；权限用页面层 `PermissionGate`。

### 7.3 共享组件索引

| 组件                                              | 路径                 | 消费者                       |
| ------------------------------------------------- | -------------------- | ---------------------------- |
| `PageShell`、`DataSection`                        | `components/layout/` | 全局                         |
| `PermissionGate`                                  | `components/auth/`   | 全局                         |
| `EmptyState`、`ErrorState`、`ConfirmActionDialog` | `components/ui/`     | 全局                         |
| `CredentialForm`、`SyncConfigPanel`               | `components/org/`    | workflow                     |
| `BudgetProgressCell`                              | `components/budget/` | budget、usage、platform keys |
| `AuditFilteredPage`、`AuditToolbar`               | `components/audit/`  | operations、calls            |

### 7.4 页面私有组件（Review 对照）

| 域        | 组件                                                                             | 页面                       |
| --------- | -------------------------------------------------------------------------------- | -------------------------- |
| org       | `structure-toolbar`、`structure-summary-card`、`department-tree`、`member-table` | structure                  |
| org       | `role-list`、`role-member-table`                                                 | roles                      |
| org       | `data-source-init-progress`、`import-result`、`sync-log-table`                   | data-source                |
| keys      | `my-keys-table`、`platform-key-table`、`provider-key-table`                      | mine / platform / provider |
| budget    | `budget-row`                                                                     | overview                   |
| dashboard | `cost-*` ×5                                                                      | cost                       |
| dashboard | `usage-model-chart`                                                              | usage                      |
| audit     | `call-logs-table`                                                                | calls                      |

**新组件 checklist：** 统计 route 页面 + workflow 引用 → 选目录 → 禁止业务域文件名进 `components/ui/`。

---

## 8. 常量、类型与命名

| 类型          | 位置                                                   |
| ------------- | ------------------------------------------------------ |
| 路由          | `config/routes.ts`                                     |
| 权限 key      | `lib/permission-keys.ts`（`permissions.ts` re-export） |
| 业务标签      | `lib/labels.ts` 或 `lib/{domain}-constants.ts`         |
| Workflow 文案 | `features/workflow/constants.ts`                       |
| 环境          | `config/app.ts`                                        |

类型：API → `api/types/`；页面 view model → 页面 Hook 旁导出。

| 类别         | 约定                            | 示例                    |
| ------------ | ------------------------------- | ----------------------- |
| 页面         | `{name}.tsx`                    | `structure.tsx`         |
| 页面 Hook    | `use-{name}-page.ts`            | `use-structure-page.ts` |
| 页面组件     | `{feature}-{part}.tsx`          | `cost-trend-chart.tsx`  |
| API          | `{domain}.ts` / `{resource}Api` | `memberApi`             |
| Mock handler | `mocks/handlers/{domain}.ts`    | 与 api 同域             |

---

## 9. 页面与 Hook 索引

16 个 `APP_ROUTES` 页面均已 Page → Hook → Components 三层。

| 页面                | Hook                         | 主要 components                                    |
| ------------------- | ---------------------------- | -------------------------------------------------- |
| `dashboard/cost`    | `use-cost-dashboard-page`    | `cost-*` ×5                                        |
| `dashboard/usage`   | `use-usage-dashboard-page`   | `usage-model-chart`                                |
| `org/structure`     | `use-structure-page`         | `structure-*`、`department-tree`、`member-table`   |
| `org/data-source`   | `use-data-source-page`       | `data-source-*`、`import-result`、`sync-log-table` |
| `org/roles`         | `use-roles-page`             | `role-list`、`role-member-table`                   |
| `budget/overview`   | `use-budget-overview-page`   | `budget-row`                                       |
| `budget/allocation` | `use-budget-allocation-page` | —                                                  |
| `budget/alerts`     | `use-budget-alerts-page`     | —                                                  |
| `keys/mine`         | `use-my-keys-page`           | `my-keys-table`                                    |
| `keys/approval`     | `use-approval-page`          | —                                                  |
| `keys/platform`     | `use-platform-keys-page`     | `platform-key-table`                               |
| `keys/provider`     | `use-provider-keys-page`     | `provider-key-table`                               |
| `models/list`       | `use-model-list-page`        | —                                                  |
| `models/routing`    | `use-model-routing-page`     | —                                                  |
| `audit/calls`       | `use-audit-calls-page`       | `call-logs-table`                                  |
| `audit/operations`  | `use-audit-operations-page`  | —                                                  |

**Hook 分层：** 单页 → `use-*-page.ts`；域内复用 → `routes/{domain}/hooks/`；跨域通用 → `hooks/`。

### 全局 Hook

| Hook                  | 职责                       |
| --------------------- | -------------------------- |
| `useAsyncResource`    | 异步 loading/error/refresh |
| `useFilteredResource` | 带 filter 的列表           |
| `useWorkflowRefresh`  | workflow 成功后 refresh    |
| `usePermissions`      | 权限与 readOnly            |
| `useRowHighlight`     | 行高亮                     |
| `usePageSubtitle`     | Header 副标题              |

---

## 10. Workflow、Demo 与 MSW

**Workflow 新增步骤：** `workflows/{name}.tsx` + `workflow-payloads.ts` + `workflow-definitions.tsx`；可复用 UI 放 `components/{domain}/`。

**Demo：** `features/demo/`（roles、guide、chrome）；`usePermissions` 读 demo session。

**MSW：** `mocks/handlers/{domain}.ts` 与 `api/{domain}.ts` 路径一致；mock 工具仅放 `mocks/lib/`。测试用 [`mocks/server.ts`](../apps/frontend/src/mocks/server.ts) + [`test-setup.ts`](../apps/frontend/src/test-setup.ts)。

---

## 11. 技术栈与工具链

- **别名：** `@/`（`tsconfig` + Vite）；类型用 `import type`；路由从 `react-router` 导入
- **UI：** Tailwind 4 + shadcn（`components/ui/`）；表格 TanStack Table；图表 Recharts（数据转换在页面 Hook）
- **表单：** react-hook-form；页面内嵌 → `components/{domain}/`，侧滑 → `features/workflow/workflows/`

| 命令                  | 作用                          |
| --------------------- | ----------------------------- |
| `pnpm start`          | 开发                          |
| `pnpm lint`           | ESLint + `check-conventions`  |
| `pnpm test`           | Vitest（`test-setup` 挂 MSW） |
| `pnpm build`          | `tsc -b && vite build`        |
| `pnpm build:gh-pages` | Mock 生产构建                 |

**CI：** [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) 与 [`deploy.yml`](../.github/workflows/deploy.yml) 均执行 lint + test + build。

**测试布局（7 文件 / 40 用例）：** `lib/*.test.ts` ×3、`config/routes.test.ts`、`hooks/use-filtered-resource.test.ts`、`routes/org/hooks/use-structure-page.test.tsx`、`components/auth/permission-gate.test.tsx`；工具 [`test-utils.tsx`](../apps/frontend/src/test-utils.tsx)。

**增量测试优先级：** `lib/` 纯函数 → 全局 Hook → 页面 Hook（`injectedApis`）→ 关键组件（`PermissionGate`）。

---

## 12. 反模式（Review 对照）

| 反模式                                   | 正确做法                                           |
| ---------------------------------------- | -------------------------------------------------- |
| 页面内大量 `useState` + handler          | `use-*-page.ts`                                    |
| 子组件直接 `import { xxxApi }`           | `useApis()` 或 Hook 传回调                         |
| `components/ui` 含业务语义               | `components/{domain}/` 或 `routes/.../components/` |
| 单页组件放 `components/{domain}/`        | `routes/{domain}/components/`                      |
| 硬编码路由字符串                         | `ROUTES.*`                                         |
| 新路由未更新 `APP_ROUTES` / `ROUTE_META` | 同步 config                                        |
| 忽略 `useAsyncResource` 的 `error`       | `ErrorState` 或 toast                              |
| Mock 与 api 路径不一致                   | 对照 API 契约                                      |
| mock 工具进 `src/lib/`                   | `mocks/lib/`                                       |
| 映射表多处定义                           | `lib/labels.ts` 或 constants                       |

---

## 13. 维护要点

1. **新功能：** Page → Hook → Components；路由走 `ROUTE_META` 清单；组件走 §7.4 checklist。
2. **新 API：** 契约 + handler + 类型同步；按需补测试。
3. **Review：** §12 反模式 + `pnpm lint` / `pnpm test` 本地通过后再提 PR。

架构已稳定，后续以**增量规范**为主，无需再按 Phase 路线图做大块迁移。

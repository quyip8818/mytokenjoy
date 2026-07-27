# TanStack Router 迁移方案

> 从 react-router v7 迁移到 @tanstack/react-router，获得类型安全路由、search params schema 和 loader 一体化。

---

## 1. 迁移动机

| 痛点 | react-router 现状 | TanStack Router 解法 |
|------|-------------------|---------------------|
| Deep link | 手动拼 `?tab=xxx`，无校验 | route 级声明 `validateSearch` schema，navigate 时 TS 强制传对 |
| 路径 typo | 字符串路径无编译检查 | route tree 生成 typed link，路径改了所有引用处报错 |
| 权限 guard | SessionGate + useRouteRedirect 运行时判断 | `beforeLoad` 钩子统一拦截，集中一处 |
| 数据预加载 | 组件内 useQuery，切路由时白屏 | `loader` 在路由级预加载，进入时数据已就绪 |
| Code split | 手动 `lazy(() => import(...))` + Suspense | route 自带 `lazy` 或 file-based 自动 split |

---

## 2. 现状盘点

### 2.1 依赖

- `react-router: ^7.18.0`（import from `'react-router'`，非 react-router-dom）
- `@tanstack/react-query: ^5.101.2`（已有，可复用 queryClient 做 router loader）

### 2.2 路由结构

```
App.tsx (BrowserRouter)
├── /login — LoginPage (公开)
├── /invite/accept — InviteAcceptPage (公开)
└── /* — AuthenticatedRoutes
    └── SessionGate (auth 检查)
        └── AdminLayout (sidebar + header + Outlet)
            ├── / — HomeRedirect
            ├── /dashboard/cost
            ├── /dashboard/usage
            ├── /keys/platform
            ├── /approvals
            ├── /keys/provider
            ├── /models/list
            ├── /models/routing
            ├── /budget
            ├── /budget/alerts
            ├── /billing
            ├── /org/data-source
            ├── /org/structure
            ├── /org/roles
            ├── /audit/operations
            ├── /audit/calls
            ├── /me/keys
            ├── /me/usage
            ├── /me/settings
            └── /notifications
```

### 2.3 影响面

| 类别 | 数量 |
|------|------|
| react-router import 文件 | 19 |
| useNavigate/useLocation/etc 调用 | ~39 |
| 路由定义（config/routes.ts） | 18 条 |
| NavLink 使用（sidebar） | 1 处 |
| Navigate 组件使用 | 1 处 (HomeRedirect) |
| useSearchParams 使用 | 2 处 (useUrlTab, use-dept-selection) |

---

## 3. 目标架构

### 3.1 Route Tree（code-based，不用 file-based）

保留现有 `src/routes/` 目录结构，但路由定义改为 TanStack Router 的 `createRoute` API：

```ts
// src/router/routes.ts
import { createRootRouteWithContext, createRoute, createRouter } from '@tanstack/react-router'

// Root route — 注入 queryClient 和 session context 到 router
const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: RootLayout, // AppProviders + ErrorBoundary
})

// 公开路由
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: lazy(() => import('@/routes/auth/login')),
})

// 认证布局 route
const authLayout = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  beforeLoad: ({ context }) => {
    if (!context.session) throw redirect({ to: '/login' })
  },
  component: AdminLayout, // sidebar + header + <Outlet />
})

// 业务路由（authLayout 的子路由）
const settingsRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/me/settings',
  validateSearch: z.object({ tab: z.enum(['account','security','notifications']).optional() }),
  component: lazy(() => import('@/routes/me/settings')),
})

const notificationsRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/notifications',
  component: lazy(() => import('@/routes/notifications')),
})
```

### 3.2 Search Params（Deep Link）

每个需要 deep link 的路由声明 `validateSearch`：

```ts
// settings route
validateSearch: z.object({
  tab: z.enum(['account', 'security', 'notifications']).catch('account'),
})

// notifications route（未来）
validateSearch: z.object({
  tab: z.enum(['inbox', 'archived']).catch('inbox'),
  category: z.string().optional(),
})

// keys/platform
validateSearch: z.object({
  highlight: z.string().optional(),
  projectId: z.string().optional(),
})
```

组件内使用：

```ts
// 组件内读 search params — 完全类型安全
const { tab } = Route.useSearch()

// 跳转 — TS 会校验 search 参数是否合法
navigate({ to: '/me/settings', search: { tab: 'notifications' } })
```

### 3.3 Auth Guard

从 `SessionGate` 组件迁移到 route `beforeLoad`：

```ts
const authLayout = createRoute({
  // ...
  beforeLoad: async ({ context }) => {
    // context.session 由 rootRoute 注入
    if (!context.session) {
      throw redirect({ to: '/login' })
    }
  },
})
```

权限检查：

```ts
const budgetRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/budget',
  beforeLoad: ({ context }) => {
    if (!context.permissions.includes('budget:read')) {
      throw redirect({ to: '/' })
    }
  },
})
```

### 3.4 Navigation

| Before (react-router) | After (TanStack Router) |
|----------------------|------------------------|
| `useNavigate()` → `navigate('/path')` | `useNavigate()` → `navigate({ to: '/path' })` |
| `navigate('/me/settings?tab=notifications')` | `navigate({ to: '/me/settings', search: { tab: 'notifications' } })` |
| `<NavLink to="/path">` | `<Link to="/path" activeProps={{ className: '...' }}>` |
| `<Navigate to="/path" replace />` | `<Navigate to="/path" replace />` (TanStack 版) |
| `useLocation().pathname` | `useRouterState({ select: s => s.location.pathname })` |
| `useSearchParams()` | `Route.useSearch()` (typed!) |

### 3.5 Sidebar 适配

```tsx
// Before
<NavLink to={item.path} className={({ isActive }) => cn(..., isActive && 'active')}>

// After
<Link to={item.path} activeProps={{ className: 'active' }} inactiveProps={{ className: 'inactive' }}>
```

---

## 4. 实施步骤

### Phase 1：基础设施搭建（新增，不删旧的）

1. 安装依赖：`@tanstack/react-router`、`@tanstack/react-router-devtools`（dev）、`zod`（search params validation）
2. 创建 `src/router/` 目录：
   - `context.ts` — RouterContext 类型（queryClient + session）
   - `root.tsx` — rootRoute（RootLayout）
   - `auth-layout.tsx` — authLayout route（beforeLoad 检查 session）
   - `routes.ts` — 所有业务路由定义
   - `index.ts` — createRouter + routeTree 组装 + export
3. 改写 `main.tsx`：`<RouterProvider router={router} />`
4. 删除 `App.tsx`（逻辑全搬进 router）

### Phase 2：路由逐一迁移

按 nav group 顺序迁移每组路由：
1. 迁移路由定义（config/routes.ts → router/routes.ts）
2. 迁移页面组件中的 hooks（useNavigate → TanStack 版）
3. 迁移 search params 使用处（useSearchParams → Route.useSearch()）

### Phase 3：清理

1. 删除 react-router 依赖
2. 删除旧 config/routes.ts 中路由定义（保留 NAV_GROUP_LAYOUT 给 sidebar）
3. 删除 SessionGate、useRouteRedirect、useRouteAccess（职责已移入 beforeLoad）
4. 删除 HomeRedirect（改为 index route 的 beforeLoad redirect）
5. 删除 useUrlTab（search params 由路由层 validateSearch 管理）

---

## 5. 文件结构（迁移后）

```
src/
├── router/
│   ├── context.ts          # RouterContext type
│   ├── root.tsx            # rootRoute + RootLayout
│   ├── auth-layout.tsx     # authenticated layout route
│   ├── routes/
│   │   ├── dashboard.ts    # /dashboard/* routes
│   │   ├── keys.ts         # /keys/* routes
│   │   ├── models.ts       # /models/* routes
│   │   ├── budget.ts       # /budget/* routes
│   │   ├── org.ts          # /org/* routes
│   │   ├── audit.ts        # /audit/* routes
│   │   ├── me.ts           # /me/* routes
│   │   ├── notifications.ts
│   │   └── auth.ts         # /login, /invite/accept
│   └── index.ts            # createRouter, export router
├── routes/                 # 页面组件（保持不变）
├── features/               # 业务逻辑（保持不变）
└── main.tsx                # RouterProvider 入口
```

---

## 6. 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| code-based vs file-based | code-based | 与现有 lazy import 对齐，路由 ~20 条不需要 file convention |
| search params validation | Zod | 项目已用 Zod 做 form validation，不引新依赖 |
| 是否用 router loader | 暂不用 | 现有 TanStack Query 模式运转良好，不需要改数据获取层 |
| DevTools | dev only | `@tanstack/react-router-devtools` 方便调试 route tree |
| 权限 guard 位置 | `beforeLoad` | 集中一处，比组件内检查更早生效 |
| Sidebar nav 数据源 | 保留 `NAV_GROUP_LAYOUT` | nav 配置与路由解耦，sidebar 不需要知道 route 细节 |

---

## 7. 风险与注意

1. **react-router v7 和 TanStack Router 不能共存** — 需要一次性完整迁移，不能渐进
2. **SSE 连接不受路由切换影响** — NotificationProvider 在 rootRoute layout 层，不会因路由变化重连
3. **Workflow panel（overlay）** — 在 rootRoute layout 层渲染，不受路由影响
4. **embed.html 入口** — 确认 embed 页面是否需要路由（如果只是 iframe 单页，可以独立 entry 不走 router）

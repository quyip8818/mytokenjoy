# TanStack Router 迁移方案（已完成）

> 从 react-router v7 迁移到 @tanstack/react-router，获得类型安全路由、search params schema 和 code split 一体化。

---

## 1. 迁移动机

| 痛点 | react-router 现状 | TanStack Router 解法 |
|------|-------------------|---------------------|
| Deep link | 手动拼 `?tab=xxx`，无校验 | route 级声明 `validateSearch` schema，navigate 时 TS 强制传对 |
| 路径 typo | 字符串路径无编译检查 | route tree 生成 typed link，路径改了所有引用处报错 |
| 权限 guard | SessionGate + useRouteRedirect 运行时判断 | AuthenticatedLayout 内 permission watcher 集中处理 |
| Code split | 手动 `lazy(() => import(...))` + Suspense | route 自带 `lazyRouteComponent` 自动 split |

---

## 2. 最终架构

### 2.1 组件层次

```
main.tsx — RouterProvider
└── rootRoute (RootLayout)
    └── AppProviders (ApiProvider > QueryProvider > AuthSessionProvider > NotificationProvider)
        ├── /login — LoginPage (公开)
        ├── /invite/accept — InviteAcceptPage (公开)
        └── authLayoutRoute (AuthenticatedLayout)
            └── SessionGate 逻辑 + AdminLayout + permission watcher
                ├── / — HomeRedirect
                ├── /dashboard/cost (validateSearch: { dept })
                ├── /dashboard/usage (validateSearch: { dept })
                ├── /keys/platform (validateSearch: { highlight, projectId })
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
                ├── /me/settings (validateSearch: { tab })
                └── /notifications
```

### 2.2 Session & Auth 策略

Session 保留为 **React Context**（`AuthSessionProvider` + TanStack Query），未重构为 zustand store。

认证和权限检查在 **组件级**（`AuthenticatedLayout` 内），不使用 `beforeLoad`：
- 401 检测：`useEffect` + `window.location.replace(LOGIN_PATH)`
- 权限 watcher：`useEffect` 监听 `permissions` + `pathname` 变化，失权时 redirect

```tsx
// router/auth-layout.tsx
function AuthenticatedLayout() {
  const { sessionError, loading, permissions, refreshSession } = useSession()
  const router = useRouter()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  // 401 → redirect to login
  useEffect(() => {
    if (isUnauthorized) window.location.replace(LOGIN_PATH)
  }, [isUnauthorized])

  // Permission watcher
  useEffect(() => {
    if (!canAccessRoute(pathname, permissions)) {
      void router.navigate({ to: getDefaultHomePath(permissions) ?? '/', replace: true })
    }
  }, [permissions, pathname])

  // ... render sidebar + header + Outlet
}
```

**升级路径**：session 改 zustand 后可以把 auth 检查移到 `beforeLoad`。

### 2.3 Search Params（validateSearch）

| 路由 | Schema |
|------|--------|
| `/me/settings` | `z.object({ tab: z.enum(['account','security','notifications']).catch('account') })` |
| `/dashboard/cost` | `z.object({ dept: z.string().optional() })` |
| `/dashboard/usage` | `z.object({ dept: z.string().optional() })` |
| `/keys/platform` | `z.object({ highlight: z.string().optional(), projectId: z.string().optional() })` |
| `/invite/accept` | `z.object({ code: z.string().optional() })` |

组件内读取方式：
- `/me/settings` 用 `useSearch({ strict: false })` 读 `tab`
- `/dashboard/*` 用 `useRouterState({ select: s => s.location.search })` 读 `dept`
- `/invite/accept` 用 `useSearch({ strict: false })` 读 `code`

### 2.4 Navigation API

| Before (react-router) | After (TanStack Router) |
|----------------------|------------------------|
| `navigate('/path')` | `navigate({ to: '/path' })` |
| `navigate('/path', { replace: true })` | `navigate({ to: '/path', replace: true })` |
| `navigate('/me/settings?tab=x')` | `navigate({ to: '/me/settings', search: { tab: 'x' } })` |
| `<NavLink to={path} className={({ isActive }) => ...}>` | `<Link to={path}>` + 手动 `useRouterState` 判断 isActive |
| `<Link to="/path">` | `<Link to="/path">` |
| `<Navigate to="/path" replace />` | `<Navigate to="/path" replace />` |
| `useLocation().pathname` | `useRouterState({ select: s => s.location.pathname })` |
| `useSearchParams()` | `useRouterState` 或 `useSearch({ strict: false })` |

### 2.5 Code Split

使用 `lazyRouteComponent`：

```ts
const budgetRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/budget',
  component: lazyRouteComponent(() => import('@/routes/budget')),
})
```

### 2.6 NotificationProvider 位置

`NotificationProvider` 在 **rootRoute** 层（`AppProviders` 内，`AuthSessionProvider` 之后）。
SSE 连接由 `useNotificationConnection` 在组件内条件启动，login 页面不会触发。

---

## 3. 文件结构

```
src/
├── router/
│   ├── context.ts          # RouterContext type（暂为空，预留 loader）
│   ├── root.tsx            # rootRoute + RootLayout（AppProviders 包装）
│   ├── auth-layout.tsx     # authLayoutRoute（session 检查 + AdminLayout + permission watcher）
│   ├── routes/
│   │   ├── home.tsx        # / → HomeRedirect（权限决定默认页）
│   │   ├── dashboard.ts    # /dashboard/* routes
│   │   ├── keys.ts         # /keys/*, /approvals
│   │   ├── models.ts       # /models/* routes
│   │   ├── budget.ts       # /budget/*, /billing
│   │   ├── org.ts          # /org/* routes
│   │   ├── audit.ts        # /audit/* routes
│   │   ├── me.ts           # /me/* routes
│   │   ├── notifications.ts
│   │   └── auth.ts         # /login, /invite/accept
│   └── index.ts            # createRouter + routeTree + Register 声明
├── routes/                 # 页面组件（保持不变）
├── features/               # 业务逻辑（保持不变）
├── config/routes.ts        # ROUTE_DEFINITIONS, NAV_GROUP_LAYOUT, 权限工具
└── main.tsx                # RouterProvider 入口
```

---

## 4. 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| code-based vs file-based | code-based | 路由 ~20 条，不需要 file convention |
| search params validation | Zod | 类型推导优秀，社区标配 |
| lazy 方式 | `lazyRouteComponent` | 与现有按页面 split 模式一致 |
| 是否用 router loader | 暂不用 | TanStack Query 模式不动 |
| session 注入方式 | React Context（保持现状） | 避免 session 层重构，auth 检查在组件级 |
| 认证检查位置 | AuthenticatedLayout 组件内 | session 来自 React Context，无法在 beforeLoad 同步读取 |
| 权限动态变化 | AuthenticatedLayout 内 useEffect watcher | 监听 permissions + pathname 变化 |
| DevTools | dev only 安装 | `@tanstack/react-router-devtools`（未在 RootLayout 中渲染，按需启用） |
| 权限 source of truth | `ROUTE_DEFINITIONS.requiredPermissions` + `canAccessRoute()` | 集中一处 |
| Sidebar nav 数据源 | 保留 `NAV_GROUP_LAYOUT` | nav 配置与路由定义解耦 |
| 迁移策略 | 一次性切换 | 体量小，已完成 |

---

## 5. 已删除文件

| 文件 | 替代 |
|------|------|
| `App.tsx` | `router/index.ts` + `main.tsx` |
| `components/layout/admin-layout.tsx` | `router/auth-layout.tsx` |
| `components/layout/home-redirect.tsx` | `router/routes/home.tsx` |
| `features/session/session-gate.tsx` | `router/auth-layout.tsx` 内 AuthenticatedLayout |
| `features/session/session-navigation-bridge.tsx` | `router/auth-layout.tsx` 内 permission watcher |
| `features/session/use-route-redirect.ts` | `router/auth-layout.tsx` 内 permission watcher |
| `features/session/use-route-access.ts` | `canAccessRoute()` 直接在 auth-layout 调用 |
| `hooks/use-url-tab.ts` | route `validateSearch` + `useSearch` |
| `config/routes.ts` 中 `APP_ROUTES` | 不再需要（router tree 替代） |
| `config/routes.ts` 中 `toRouterPath` | 不再需要 |

---

## 6. 升级路径

1. **Session → zustand**：将 session 从 React Context 提取到 zustand store，然后 auth 检查可移到 `beforeLoad`
2. **Router loader**：需要数据预加载时，把 `queryClient` 注入 `RouterContext`，在 `loader` 中调用 `queryClient.ensureQueryData()`
3. **DevTools**：在 `RootLayout` 中 conditionally 渲染 `<TanStackRouterDevtools />`
4. **Typed navigation**：当前部分 navigate 用 `{ to: '.' }` 或 string path，未来可以逐步改用 route reference 获得完整类型检查

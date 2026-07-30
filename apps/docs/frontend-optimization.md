# Frontend 架构优化分析

> 最后更新：2026-07-30（反映 `12772876` API 层重构 + `7f111dc5` devApi 移除）

## 现状总评

架构成熟度较高：feature-based 模块化、barrel exports 封装、API DI 层（`useApis()`）、TanStack Query 管理服务端状态、声明式路由权限、lazy loading。日常开发体验良好，新增页面的脚手架成本低。

### 最近重要变化

| Commit | 内容 | 影响 |
|--------|------|------|
| `12772876` | API 层合并（20→13 个域 API），嵌套结构化（`orgApi.members.list`），account 页改用 `useMutation` | API 膨胀问题已解决；mutation 模式开始统一 |
| `7f111dc5` | 删除 `devApi`（被 `keysApi.platform.simulateBearer` 替代） | 减少无用代码 |
| `4bf27183` | 迁移到 `@tanstack/react-router` | 路由类型安全 + lazy loading 到位 |
| `b3e97f99` | 统一页面布局系统（`PageShell` + `DataSection`） | 数据加载错误处理一致性提升 |

下面聚焦**仍然存在且真正影响可扩展性和稳定性的架构风险**，按影响排序。

---

## 一、Mutation 防重复提交不一致（高优先级）

### 现状

最近重构已在 `use-account-page.ts` 中用 `useMutation`（TanStack Query 原生）替代了手动 useState，这是正确方向。但系统中仍有两种 mutation 模式并存：

**已改造（有 isPending 锁）**：
- `use-account-page.ts` — 直接用 `useMutation`（5 个 mutation 都有 `isPending`）
- `use-workflow-submit.ts` — 用 `useInjectedMutation`
- `use-notification-inbox.ts` — 用 `useInjectedMutation`

**未改造（无防重复）**：
- `use-budget-actions.ts` — 8 个 async callback（`updateDepartment`, `deleteProject` 等）
- `use-model-list-page.ts` — `handleToggle`, `handleDelete`
- `use-platform-keys-page.ts` — `handleDelete`
- 其他 page hook 中的 imperative async 操作

```ts
// use-budget-actions.ts — 仍无防重复
const deleteProject = useCallback(
  (groupId: string) =>
    withErrorToast(async () => {
      await apis.budgetApi.deleteProject(groupId)  // 可被连续触发
      await refresh()
    }, '删除项目失败'),
  [apis, refresh],
)
```

### 风险

- 预算分配、项目删除、模型切换等**有副作用操作**可被重复提交
- 后端若无幂等保护，可能产生脏数据

### 建议

参照 `use-account-page.ts` 的模式，将 budget/models/keys 的 write 操作也迁移到 `useMutation`。已有参考实现，只需跟进：

```ts
// 参照 account 的模式：
const deleteMutation = useMutation({
  mutationFn: (groupId: string) => apis.budgetApi.deleteProject(groupId),
  onSuccess: () => { void refresh(); toast.success('项目已删除') },
  onError: (err) => toast.error(apiErrorMessage(err, '删除项目失败')),
})
// deleteMutation.isPending → disable 按钮
```

**工作量**：~2h（budget-actions + model-list + keys，照搬 account 模式）。

---

## 二、API 请求不支持取消（中优先级）

### 现状

- `client.ts` 的 `request()` 函数不接受 `signal` 参数
- `useInjectedQuery` 的 `queryFn` 签名为 `(apis: AppApis) => Promise<T>`，丢弃了 TanStack Query 传入的 `{ signal }`
- 整个前端代码无一处使用 `AbortController`

### 风险

- 用户快速切换页面时，旧页面的请求仍在 flight，浪费带宽（移动端更明显）
- 搜索场景（audit/keys filter）不会取消上次请求，存在竞态——旧响应晚于新响应回来时覆盖正确数据
- TanStack Query 在组件 unmount 时会调用 abort（如果 signal 被传递），但当前配置下此机制完全失效

### 建议

两步改造：

1. `client.ts` 的 `request()` 增加 `signal?: AbortSignal` 参数，传递给 `fetch(url, { ...init, signal })`
2. `useInjectedQuery` 的 `queryFn` 签名改为 `(apis: AppApis, context: { signal: AbortSignal }) => Promise<T>`，将 TanStack Query 提供的 signal 传下去

```ts
// client.ts
export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_PATH}${path}`
  const init: RequestInit = { credentials: 'include', ...options }
  // signal 已在 options 中，无需特殊处理
  let res = await fetch(url, init)
  // ...
}

// use-injected-query.ts
queryFn: ({ signal }) => queryFn(apis, { signal }),
```

各 api 文件可渐进式接入：需要取消的场景（搜索、列表筛选）先传 signal，其他不改。

**工作量**：核心改动 30min，渐进接入各 API 按需。

---

## 三、错误隔离粒度不足（中优先级）

### 现状

错误处理分两层：
- **数据加载错误**：`DataSection` 组件做 loading/error/empty 状态切换 → 就地显示 `ErrorState` + 重试。覆盖面好，API 失败不会白屏。
- **渲染异常**：仅有 2 个 `AppErrorBoundary`（root + auth-layout 的 `<Outlet>` 外层）。feature 内部的 JS 运行时错误（解构 null、render throw）会冒泡到 auth-layout 级别，**白屏整个 main content 区域**。

Sidebar 和 Header 不受影响（在 boundary 外），但用户需要手动导航到另一个页面才能恢复。

### 风险

一个 feature 的边缘 bug（如后端返回异常数据 → 组件 crash）会影响用户对整个系统的信心。

### 建议

在 route 级别包一层 boundary。TanStack Router 支持 `errorComponent` per route：

```ts
// router/routes/budget.ts
export const budgetRoute = createRoute({
  ...
  errorComponent: RouteErrorFallback,  // 自动捕获该 route 下的渲染错误
})
```

或者更简单：在 `auth-layout.tsx` 的 `<Outlet>` 处利用 TanStack Router 的 `defaultErrorComponent`：

```ts
export const router = createRouter({
  routeTree,
  defaultErrorComponent: RouteErrorFallback,  // 每个 route 独立 boundary
})
```

这样一个页面 crash 只影响自己，不白屏邻居。

**工作量**：30min（加 `defaultErrorComponent`），若要自定义每个 route 的错误 UI 则 ~2h。

---

## 四、权限模型缺少数据范围层（不需要做）

### 现状

权限控制有两个粒度：
- **路由级**：`config/routes.ts` 声明 `requiredPermissions`，未授权路由直接不渲染/redirect
- **组件级**：`<PermissionGate permission={...}>` 隐藏 UI 元素

两者都是 **UI 可见性控制**。数据范围完全依赖后端 API 的返回范围（后端 session middleware 用 `CompanyID` + `MemberID` 过滤）。

### 结论：当前阶段不需要做

经详细分析：
1. **后端已做好隔离**：`RequireSession` middleware 从 token 提取 `CompanyID` + `MemberID`，所有查询自动限定范围。前端不需要二次过滤。
2. **唯一的 row-level 判断**：`project-detail.tsx` 用 `project.ownerId === memberId` 控制删除按钮——这是 UI 逻辑不是权限逻辑，不需要额外框架。
3. **没有混合可见性的列表**：当前所有列表要么全部可操作（管理员），要么全部只读（成员看自己的）。不存在"部分行可编辑"的需求。
4. **`readOnly` session flag** 已处理只读角色——`usePermissions().canWrite` 全局禁写。

**何时需要重新评估**：出现"部门主管看本部门+下级部门数据"的前端需求时。目前后端已支持这种过滤，前端无需改动。

---

## 五、跨 Feature 通信健康度评估（不需要做）

### 现状

Feature 间通信有 3 种模式：
1. **Workflow onSuccess 回调** → 触发 `invalidateQueries` + `refresh`（主方式）
2. **`apiEvents` 事件总线** → 仅 API 生命周期事件（unauthorized/forbidden/authzRevision），feature 不直接使用
3. **Barrel import** → 引用其他 feature 导出的纯工具/常量/类型（不含状态）

### 结论：不需要做

经详细分析跨 feature 依赖图：
- **workflow → models/org/keys/dashboard**：仅引用 labels、util 函数、类型。纯数据依赖，无状态耦合。
- **dashboard → budget**：仅引用 `findBudgetNode`、`getBudgetProgressClass` 等纯函数。
- **account → auth**：仅引用 `useVerifyCountdown`（一个独立 timer hook）。
- **account → notifications**：引用 `useNotificationsPage`、`NotificationsPageShell`（settings 页面内嵌通知设置 tab）。

所有跨 feature 引用都是：
- 纯函数/常量/类型（无状态泄漏）
- 通过 barrel 导入（不违反封装）
- 单向依赖（无循环）

**没有** feature-to-feature 的全局 store 共享、事件广播、或隐式耦合。TanStack Query 的 invalidation 机制是唯一的数据同步手段——声明式、可追踪、无副作用。

**不需要引入 event bus 或其他通信机制。**

---

## 六、已完成的优化（本轮）

| 项目 | 状态 | Commit |
|------|------|--------|
| Mutation 防重复（budget/models/keys） | ✅ 已完成 | `c9ca33e9` |
| API signal 支持（useInjectedQuery 传递 signal） | ✅ 已完成 | `c9ca33e9` |
| 路由级错误隔离（defaultErrorComponent） | ✅ 已完成 | `c9ca33e9` |
| useAsyncFetch 收口到 useQuery | ✅ 已完成 | `3e8d103b` |
| 权限数据范围 | ✅ 不需要做 | — |
| 跨 feature 通信 | ✅ 不需要做 | — |

---

## 不需要动的

- **Barrel 封装 + self-barrel import**：不产生真实循环依赖，Vite 处理正确。代码审美问题，不是架构风险。
- **API 层组织**：最近重构已将 20 个平铺 API 合并为 13 个嵌套结构（`orgApi.members.list`），清晰且 type-safe。不再膨胀。
- **`config/routes.ts` 250 行**：单一概念的数据文件，拆开反而增加认知跳转。
- **Workflow 作为单一 feature**：只有一个挂载点（auth-layout），不满足"通用基础设施需要多消费者"的分层前提。
- **TanStack Query invalidation 机制**：已是前端数据同步的最佳实践。
- **DI 模式（`useApis()`）**：测试可替换，AppApis 接口字段已精简到 13 个，注册成本低。

---

## 附录：API 层重构后的架构图

```
src/api/
├── app-apis.ts       ← AppApis interface (13 fields)
├── client.ts         ← request() + buildQuery() + ApiError
├── api-events.ts     ← typed event bus (unauthorized/forbidden/authzRevision)
├── context.tsx       ← ApiProvider
├── use-apis.ts       ← useApis() / useInjectedApis()
│
├── approval.ts       ← approvalApi
├── audit.ts          ← auditApi
├── auth.ts           ← authApi
├── billing.ts        ← billingApi (nested: invoices)
├── budget.ts         ← budgetApi
├── dashboard.ts      ← dashboardApi
├── keys.ts           ← keysApi { provider: {...}, platform: {...} }
├── me.ts             ← meApi
├── models.ts         ← modelsApi { list, create, ..., routing: {...} }
├── notifications.ts  ← notificationApi
├── org.ts            ← orgApi { dataSource, sync, departments, members, roles }
├── platform.ts       ← platformApi
├── session.ts        ← sessionApi
├── setup.ts          ← setupApi
│
└── types/            ← DTO 类型（按域拆分）
    ├── approval.ts
    ├── audit.ts
    ├── billing.ts
    ├── budget.ts
    ├── common.ts
    ├── dashboard.ts
    ├── index.ts
    ├── keys.ts
    ├── me.ts
    ├── models.ts
    ├── mydashboard.ts
    ├── notification.ts
    ├── org.ts
    └── platform.ts
```

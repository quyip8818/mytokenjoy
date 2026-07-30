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

## 四、权限模型缺少数据范围层（低优先级，需观察）

### 现状

权限控制有两个粒度：
- **路由级**：`config/routes.ts` 声明 `requiredPermissions`，未授权路由直接不渲染/redirect
- **组件级**：`<PermissionGate permission={...}>` 隐藏 UI 元素

两者都是 **UI 可见性控制**。数据范围完全依赖后端 API 的返回范围（如 `memberId` 筛选）。

### 当前是否够用？

对于本系统（企业内部 LLM API 管控台），当前模型**暂时够用**：
- 大部分列表是"管理员看全部"或"成员看自己"——后端按 session 中的身份返回对应范围
- 没有"A 部门主管只能看本部门数据"的细粒度场景（或有但完全由后端处理）

### 何时会痛

如果出现以下需求，前端需要加数据范围感知：
- 列表中混合展示"可操作"和"只读"行（需要 row-level permission badge）
- 前端缓存了全量数据 + 客户端过滤（当前未做，都是服务端过滤）
- 需要 optimistic update 但只针对有权限的行

### 建议

- 短期：保持现状。前端**不**做数据过滤，相信后端返回范围。在 API 层文档化这个约定。
- 长期：如需行级权限展示，在 `api/types` 中让后端返回 `canEdit` / `canDelete` 字段（row-level capability），前端按此渲染操作按钮。不在前端做权限计算。

---

## 五、跨 Feature 通信健康度评估（低优先级）

### 现状

Feature 间通信有 3 种模式：
1. **Workflow onSuccess 回调** → 触发 `invalidateQueries` + `refresh`（主方式）
2. **`apiEvents` 事件总线** → 仅 API 生命周期事件（unauthorized/forbidden/authzRevision），feature 不直接使用
3. **Barrel import** → 引用其他 feature 导出的纯工具/常量/类型（不含状态）

### 评估

这个模式**很健康**：
- 没有 feature-to-feature 的状态共享或事件广播
- 数据同步全靠 TanStack Query 的 invalidation（声明式、可追踪）
- Workflow 是唯一的"协调者"角色，但其耦合是通过 callback + query key 实现，不是硬依赖

### 唯一注意点

`useWorkflowRefresh` 的 `invalidateKeys` 参数由调用方传入（如 `queryKeys.keys.all`）。这意味着"workflow 完成后刷新哪些数据"的知识分散在各 feature 的 page hook 中——这是**正确的**做法（知识在消费者处），不需要集中化。

**结论**：不需要改动。

---

## 六、`useAsyncFetch` 应收口到 TanStack Query（低优先级）

### 现状

`features/budget/hooks/use-async-fetch.ts` 是一个手工 fetch hook：
- 用 `cancelled` flag 防止 unmount 后 setState
- 无 cache、无 dedup、无 retry
- 本质是 TanStack Query 的退化版

当前被 2 处使用：
- `budget-edit-member-budget.tsx`（对话框内按需加载成员预算）
- `use-member-budgets.ts`（部门成员预算列表）

### 建议

用 `useInjectedQuery` + `enabled` 参数替换。传入动态 key 即可实现"按需加载"语义。

**工作量**：30min，影响范围小。

---

## 优先级总结

| 优先级 | 问题 | 影响 | 工作量 |
|--------|------|------|--------|
| **高** | Mutation 无 isPending 锁 | 重复提交 → 脏数据 | 2h |
| **中** | API 不支持 abort/signal | 竞态 + 带宽浪费 | 30min 核心 + 渐进 |
| **中** | 错误边界只有 layout 级 | 一个页面 crash 白屏全 app | 30min |
| 低 | 权限无数据范围层 | 当前够用，未来可能需要 | 按需 |
| 低 | useAsyncFetch 未收口 | 维护负担轻微 | 30min |
| ✓ | 跨 feature 通信 | 健康，无需改动 | 0 |

---

## 不需要动的

- **Barrel 封装 + self-barrel import**：不产生真实循环依赖，Vite 处理正确。代码审美问题，不是架构风险。
- **API 层组织**：最近重构已将 20 个平铺 API 合并为 13 个嵌套结构（`orgApi.members.list`），清晰且 type-safe。不再膨胀。
- **`config/routes.ts` 250 行**：单一概念的数据文件，拆开反而增加认知跳转。
- **Workflow 作为单一 feature**：只有一个挂载点（auth-layout），不满足"通用基础设施需要多消费者"的分层前提。
- **TanStack Query invalidation 机制**：已是前端数据同步的最佳实践。
- **DI 模式（`useApis()`）**：测试可替换，AppApis 接口字段已精简到 13 个，注册成本低。

---

## 执行建议

1. **本周**：budget/models/keys 的 write 操作迁移到 `useMutation`（照搬 account 模式）+ `defaultErrorComponent`（合计 ~3h）
2. **近期**：`request()` 接入 signal，搜索/筛选场景先用起来；`useAsyncFetch` 改为 `useInjectedQuery`
3. **观察**：权限数据范围按需演进，跨 feature 通信保持现状

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

# data-testid 迁移：PageHeader 标记 + e2e 重写

PageHeader 组件新增可选 `testId` prop，sidebar 导航按钮/链接加上 `data-testid`，18 个 e2e 文件从 `getByRole('banner').getByRole('heading')` 改为 `getByTestId('page-*')`。同时删除了若干重复或无意义的测试用例。

Watch for:
- **7 个 testId 在源码中不存在**（confirmed）—— `page-org-structure`、`page-org-roles`、`page-budget`、`page-keys-platform`、`page-dashboard-cost`、`page-dashboard-usage`、`page-models-routing` 被 e2e 引用但对应 page shell 没有添加 testId，涉及的 smoke 和专题测试将全部失败。
- **testId 错配**（confirmed）—— `sms-sync-models.spec.ts` 在 `/models/list` 路由等待 `page-platform-models`，但该页面实际 testId 是 `page-models-list`。
- **discount.spec.ts 断言语义变化**（likely）—— 对 `[data-testid^="platform-companies-discount-"]` 的 `toBeAttached()` 断言仅证明 DOM 节点存在，不再验证下拉菜单可交互。

## High-level view

迁移把 e2e 的页面就绪检查从角色选择器 (`getByRole('heading')`) 切换到确定性的 `data-testid` 锚点，意图正确且基础设施（`PageHeader.testId`、sidebar `data-testid` 生成规则）已到位。但实施不完整：16 个 page shell 中只有 9 个真正添加了 testId，剩下 7 个 SplitPanel 布局页面被遗漏。结果是 smoke.spec.ts 里 16 条路由中有 7 条会 timeout 失败，加上若干专题测试文件的 `beforeEach` 同样会挂。

`sms-sync-models.spec.ts` 引用了错误的 testId（`page-platform-models` vs `page-models-list`），属于笔误。`discount.spec.ts` 对下拉菜单的断言从"点击打开并验证 menuitem"降级为"toBeAttached"，覆盖率有所下降但不算 blocker。

<details>
<summary>Issues (3)</summary>

1. **7 个 page shell 缺 testId** — `StructurePageShell`、`RolesPageShell`、`BudgetPageShell`、`PlatformKeysPageShell`、`CostDashboardLayoutPageShell`、`UsageDashboardLayoutPageShell`、`ModelRoutingPageShell` 需要在顶层容器加 `data-testid`（这些页面不用 PageHeader，需直接在 PageShell 或其 wrapper div 上加）。
2. **sms-sync-models testId 笔误** — 将 `page-platform-models` 改为 `page-models-list`，或确认测试意图是验证 platform 路由。
3. **discount 下拉菜单覆盖降级** — `toBeAttached()` 只验证 DOM 存在，不验证交互。考虑恢复 click + assert menuitem 的断言，用 testId 定位 trigger 即可。

</details>

<details>
<summary>Details</summary>

## 7 个 SplitPanel 页面缺少 testId

这些页面共同特征：不使用 `PageHeader`，而是 `PageShell` + `SplitPanel` 布局。迁移只覆盖了"有 PageHeader 的页面"，遗漏了所有 SplitPanel-first 的页面。

| 测试期望的 testId | 对应 shell 文件 |
|---|---|
| `page-org-structure` | `structure-page-shell.tsx` |
| `page-org-roles` | `roles-page-shell.tsx` |
| `page-budget` | `budget-page-shell.tsx` |
| `page-keys-platform` | `platform-keys-page-shell.tsx` |
| `page-dashboard-cost` | `cost-dashboard-layout-page-shell.tsx` |
| `page-dashboard-usage` | `usage-dashboard-layout-page-shell.tsx` |
| `page-models-routing` | `model-routing-page-shell.tsx` |

修复方式：`PageShell` 目前只透传 `className`，需要支持 `data-testid`（加一个 `testId` prop 或透传 rest props），然后在这 7 个 shell 传入对应 testId。

涉及测试文件：`smoke.spec.ts`（7 条）、`navigation.spec.ts`（1 条）、`org-structure.spec.ts`（8 条 beforeEach）、`org-roles.spec.ts`（7 条 beforeEach）、`budget.spec.ts`（2 条）、`budget-org-member-picker.spec.ts`（1 条）、`dashboard-cost.spec.ts`（2 条）、`feishu-import.spec.ts`（1 条）、`member-delete-no-accumulate.spec.ts`（1 条）。

## sms-sync-models 路由与 testId 错配

```typescript
// sms-sync-models.spec.ts
await page.goto('/models/list')
await expect(page.getByTestId('page-platform-models')).toBeVisible()
```

`/models/list` 渲染 `ModelListPageShell`，其 testId 是 `page-models-list`。`page-platform-models` 属于 `/platform/models` 路由的 `PlatformModelsPageShell`。

## discount 断言降级

旧代码：
```typescript
const moreBtn = page.locator('button[class*="h-8 w-8"]').first()
if (await moreBtn.isVisible()) {
  await moreBtn.click()
  await expect(page.getByRole('menuitem', { name: '优惠' })).toBeVisible()
}
```

新代码：
```typescript
const firstMoreBtn = page.locator('[data-testid^="platform-companies-discount-"]').first()
await expect(firstMoreBtn).toBeAttached()
```

新选择器更稳定（用 testId 替代 class 选择器），但断言从"打开菜单并验证可见 menuitem"变成"DOM 节点存在"。`DropdownMenuItem` 如果变为 portal 渲染或 lazy mount，这个断言不会捕捉到回归。建议在 testId selector 的基础上恢复 click → assert menuitem 的流程。

</details>

<details>
<summary>File map</summary>

- `apps/frontend/src/components/layout/page-header.tsx` — 新增 `testId?: string` prop，渲染为 `data-testid`
- `apps/frontend/src/components/layout/sidebar.tsx` — nav group 按钮加 `nav-group-{name}` testId，nav item 链接加 `nav-{path}` testId
- `apps/frontend/src/features/*/components/*-page-shell.tsx` (x14) — 向 PageHeader 传 testId
- `apps/frontend/src/features/platform/companies/components/discount-sheet.tsx` — SheetContent + Button 加 testId
- `apps/frontend/src/features/platform/companies/components/platform-companies-page-shell.tsx` — DropdownMenuItem 加动态 testId
- `apps/frontend/e2e/*.spec.ts` (x18) — 从 heading 选择器迁移到 getByTestId

完整 diff: `git diff fc980d85..HEAD`

</details>

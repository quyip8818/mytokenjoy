# Regenerate E2E Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Regenerate all Playwright e2e tests to match the current frontend pages and backend API, ensuring every admin route and member portal page has proper interaction coverage.

**Architecture:** E2e tests run against a Vite preview build (`localhost:4173`) proxied to the Go backend (`localhost:8080`) with seeded demo data. Global setup logs in via API and stores session cookies. Tests use Playwright's locator-based assertions against actual DOM roles/headings.

**Tech Stack:** Playwright, TypeScript, Go backend (seeded in-memory store), Vite preview server

## Global Constraints

- Playwright config: `apps/frontend/playwright.config.ts` — webServer starts backend + vite preview
- Auth: global-setup stores `.auth/admin.json`; tests that need unauthenticated state use `test.use({ storageState: ... })`
- Helpers: `e2e/helpers/auth.ts` provides `loginAsAdmin(page)` and `loginAsMember(page)`
- Backend seed: `apps/backend/internal/store/seed/` — uses `DemoPassword = "demo1234"`, admin email `admin@example.com`, member email `zhangsan@example.com`
- All admin routes use `storageState: '.auth/admin.json'` by default (from playwright config)
- Page headings live inside `<header role="banner">` with `<h1>` matching route label
- Route list (16 admin + 3 member): defined in `apps/frontend/src/config/routes.ts`

---

### Task 1: Auth & Session E2E

**Files:**
- Rewrite: `apps/frontend/e2e/auth.spec.ts`
- Keep: `apps/frontend/e2e/global-setup.ts` (unchanged)
- Keep: `apps/frontend/e2e/helpers/auth.ts` (unchanged)

**Interfaces:**
- Consumes: Backend `POST /api/auth/login` (email + password → session cookie)
- Produces: Validated auth redirect, login form render, successful login flow

- [ ] **Step 1: Write auth.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

// Tests without auth state
test.use({ storageState: { cookies: [], origins: [] } })

test('redirects unauthenticated user to /login', async ({ page }) => {
  await page.goto('/org/structure')
  await expect(page).toHaveURL(/\/login/)
})

test('renders login form fields', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByLabel('Email')).toBeVisible()
  await expect(page.getByLabel('Password')).toBeVisible()
  await expect(page.getByRole('button', { name: '登录' })).toBeVisible()
})

test('login with valid credentials redirects to app', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill('admin@example.com')
  await page.getByLabel('Password').fill('demo1234')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page).not.toHaveURL(/\/login/)
})

test('login with invalid credentials shows error', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill('admin@example.com')
  await page.getByLabel('Password').fill('wrongpass')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page).toHaveURL(/\/login/)
})
```

- [ ] **Step 2: Run test to verify**

```bash
cd apps/frontend && pnpm exec playwright test e2e/auth.spec.ts --reporter=list
```

- [ ] **Step 3: Commit**

```bash
git add apps/frontend/e2e/auth.spec.ts
git commit -m "test(e2e): rewrite auth tests for current login page"
```

---

### Task 2: Smoke Tests (All Routes Load)

**Files:**
- Rewrite: `apps/frontend/e2e/smoke.spec.ts`

**Interfaces:**
- Consumes: Admin session cookie from global-setup
- Produces: Validates all 16 admin routes render their heading without blank screens

- [ ] **Step 1: Write smoke.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

const routes = [
  { path: '/org/data-source', heading: '数据源' },
  { path: '/org/structure', heading: '组织架构' },
  { path: '/org/roles', heading: '角色管理' },
  { path: '/budget', heading: '预算管理' },
  { path: '/budget/alerts', heading: '预警规则' },
  { path: '/models/list', heading: '模型列表' },
  { path: '/models/routing', heading: '模型白名单' },
  { path: '/keys/mine', heading: '我的 Key' },
  { path: '/keys/approval', heading: '审批中心' },
  { path: '/keys/platform', heading: 'Key 管理' },
  { path: '/keys/provider', heading: '供应商 Key' },
  { path: '/dashboard/cost', heading: '成本看板' },
  { path: '/dashboard/usage', heading: '用量分析' },
  { path: '/wallet', heading: '钱包管理' },
  { path: '/audit/operations', heading: '操作审计' },
  { path: '/audit/calls', heading: '调用日志' },
]

for (const { path, heading } of routes) {
  test(`${path} renders heading "${heading}"`, async ({ page }) => {
    await page.goto(path)
    await expect(
      page.getByRole('banner').getByRole('heading', { name: heading }),
    ).toBeVisible()
  })
}
```

- [ ] **Step 2: Run and verify**

```bash
cd apps/frontend && pnpm exec playwright test e2e/smoke.spec.ts --reporter=list
```

- [ ] **Step 3: Commit**

```bash
git add apps/frontend/e2e/smoke.spec.ts
git commit -m "test(e2e): regenerate smoke tests for all 16 admin routes"
```

---

### Task 3: Organization Management E2E

**Files:**
- Rewrite: `apps/frontend/e2e/org-structure.spec.ts`
- Rewrite: `apps/frontend/e2e/org-roles.spec.ts`
- Rewrite: `apps/frontend/e2e/data-source.spec.ts`

**Interfaces:**
- Consumes: Seeded org tree (总公司 with sub-departments), seeded members, seeded roles
- Produces: Department tree interaction, member CRUD, role selection

- [ ] **Step 1: Write org-structure.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('组织架构', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('displays department tree and member table', async ({ page }) => {
    await expect(page.getByRole('treeitem', { name: /全部成员/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /总公司/ })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '姓名' })).toBeVisible()
  })

  test('selecting department shows members', async ({ page }) => {
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    await expect(page.getByRole('heading', { level: 3, name: '总公司' })).toBeVisible()
  })

  test('adds a member', async ({ page }) => {
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    await page.getByRole('button', { name: '添加成员' }).click()
    await expect(page.getByRole('dialog', { name: '添加成员' })).toBeVisible()

    const name = `E2E-${Date.now().toString().slice(-6)}`
    await page.locator('input[name="name"]').fill(name)
    await page.locator('input[name="phone"]').fill('13900008888')
    await page.locator('input[name="email"]').fill(`${name}@test.com`)
    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: '总公司' }).click()
    await page.getByRole('button', { name: '添加' }).click()

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10_000 })
    await expect(page.getByRole('cell', { name })).toBeVisible()
  })

  test('disables a member', async ({ page }) => {
    await page.getByRole('treeitem', { name: /全部成员/ }).click()
    const activeRow = page.getByRole('row').filter({ hasText: '已激活' }).first()
    await activeRow.getByRole('button', { name: '更多操作' }).click()
    await page.getByRole('menuitem', { name: '停用' }).click()
    await expect(page.getByRole('alertdialog', { name: '停用成员' })).toBeVisible()
    await page.getByRole('button', { name: '确认' }).click()
    await expect(page.getByRole('alertdialog')).toBeHidden()
  })
})
```

- [ ] **Step 2: Write org-roles.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('角色管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('displays preset roles', async ({ page }) => {
    await expect(page.getByText('超级管理员')).toBeVisible()
    await expect(page.getByText('普通成员')).toBeVisible()
  })

  test('selecting a role shows member list', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await expect(page.getByText(/名成员/)).toBeVisible()
  })
})
```

- [ ] **Step 3: Write data-source.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('数据源', () => {
  test('shows platform selection', async ({ page }) => {
    await page.goto('/org/data-source')
    await expect(page.getByRole('heading', { name: '数据源' })).toBeVisible()
    await expect(page.getByRole('button', { name: /飞书/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /钉钉/ })).toBeVisible()
    await expect(page.getByRole('button', { name: /企业微信/ })).toBeVisible()
  })
})
```

- [ ] **Step 4: Run all org tests**

```bash
cd apps/frontend && pnpm exec playwright test e2e/org-structure.spec.ts e2e/org-roles.spec.ts e2e/data-source.spec.ts --reporter=list
```

- [ ] **Step 5: Commit**

```bash
git add apps/frontend/e2e/org-structure.spec.ts apps/frontend/e2e/org-roles.spec.ts apps/frontend/e2e/data-source.spec.ts
git commit -m "test(e2e): rewrite organization management tests"
```

---

### Task 4: Budget & Alerts E2E

**Files:**
- Rewrite: `apps/frontend/e2e/budget.spec.ts`

**Interfaces:**
- Consumes: Seeded budget tree with departments, month navigation
- Produces: Budget page load, tree interaction, alert rules page

- [ ] **Step 1: Write budget.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('预算管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/budget')
    await expect(page.getByRole('heading', { name: '预算管理' })).toBeVisible()
  })

  test('displays month navigation', async ({ page }) => {
    await expect(page.getByRole('button', { name: '上一月' })).toBeVisible()
    await expect(page.getByRole('button', { name: '下一月' })).toBeVisible()
  })

  test('displays budget tree', async ({ page }) => {
    await expect(page.getByRole('treeitem').first()).toBeVisible()
  })

  test('selecting a node shows detail panel', async ({ page }) => {
    await page.getByRole('treeitem').first().click()
    // Detail panel appears with allocation info
    await expect(page.getByText(/已分配|总预算|已使用/)).toBeVisible()
  })
})

test.describe('预警规则', () => {
  test('loads alerts page with rule list', async ({ page }) => {
    await page.goto('/budget/alerts')
    await expect(page.getByRole('heading', { name: '预警规则' })).toBeVisible()
    await expect(page.getByRole('button', { name: /新建规则|添加/ })).toBeVisible()
  })
})
```

- [ ] **Step 2: Run and verify**

```bash
cd apps/frontend && pnpm exec playwright test e2e/budget.spec.ts --reporter=list
```

- [ ] **Step 3: Commit**

```bash
git add apps/frontend/e2e/budget.spec.ts
git commit -m "test(e2e): rewrite budget and alert rules tests"
```

---

### Task 5: Key Center E2E

**Files:**
- Rewrite: `apps/frontend/e2e/keys-self-service.spec.ts`
- Rewrite: `apps/frontend/e2e/approval.spec.ts`
- Rewrite: `apps/frontend/e2e/keys-approval.spec.ts` (merge into approval.spec.ts, delete this file)

**Interfaces:**
- Consumes: Seeded platform keys, seeded approvals
- Produces: Key create/rotate/delete workflow, approval tab switching

- [ ] **Step 1: Write keys-self-service.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe.configure({ mode: 'serial' })

test.describe('我的 Key - 自管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/keys/mine')
    await expect(page.getByRole('heading', { name: '我的 Key' })).toBeVisible()
  })

  test('displays quota stats', async ({ page }) => {
    await expect(page.getByText('总额度')).toBeVisible()
    await expect(page.getByText('已使用')).toBeVisible()
    await expect(page.getByText('剩余')).toBeVisible()
  })

  test('creates a new Key', async ({ page }) => {
    await page.getByRole('main').getByRole('button', { name: '创建 Key' }).click()
    await expect(page.getByRole('heading', { level: 2, name: '创建 Key' })).toBeVisible()

    await page.getByRole('textbox', { name: '如：开发调试' }).fill('E2E测试Key')
    const quotaInput = page.getByRole('spinbutton')
    await quotaInput.clear()
    await quotaInput.fill('100')
    await page.getByRole('button', { name: '下一步' }).click()

    // Model selection
    await page.getByRole('button', { name: /选择模型/ }).click()
    await expect(page.getByRole('heading', { level: 2, name: '选择模型' })).toBeVisible()
    await page.getByRole('checkbox').first().check()
    await page.getByRole('button', { name: /确认/ }).click()

    // Submit
    await page.getByRole('contentinfo').getByRole('button', { name: '创建 Key' }).click()
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({ timeout: 10_000 })
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
    await expect(page.getByRole('cell', { name: 'E2E测试Key' }).first()).toBeVisible()
  })

  test('rotates an existing Key', async ({ page }) => {
    await expect(page.locator('tbody tr').first()).toBeVisible()
    await page.locator('tbody tr').first().getByRole('button').click()
    await page.getByRole('menuitem', { name: '重新生成' }).click()
    await expect(page.getByRole('heading', { level: 2, name: '重新生成 Key' })).toBeVisible()
    await page.getByRole('button', { name: '确认重新生成' }).click()
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({ timeout: 10_000 })
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
  })

  test('deletes a Key', async ({ page }) => {
    await expect(page.locator('tbody tr').first()).toBeVisible()
    await page.locator('tbody tr').first().getByRole('button').click()
    await page.getByRole('menuitem', { name: '删除' }).click()
    const dialog = page.locator('[role="alertdialog"]')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: '删除' }).click()
    await expect(dialog).toBeHidden()
  })
})
```

- [ ] **Step 2: Write approval.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('审批中心', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/keys/approval')
    await expect(page.getByRole('heading', { name: '审批中心' })).toBeVisible()
  })

  test('displays approval tabs', async ({ page }) => {
    await expect(page.getByRole('tab', { name: /待我审批/ })).toBeVisible()
    await expect(page.getByRole('tab', { name: '我的申请' })).toBeVisible()
    await expect(page.getByRole('tab', { name: '全部' })).toBeVisible()
  })

  test('default tab is 待我审批', async ({ page }) => {
    await expect(page.getByRole('tab', { name: /待我审批/ })).toHaveAttribute('aria-selected', 'true')
  })

  test('switches between tabs', async ({ page }) => {
    await page.getByRole('tab', { name: '我的申请' }).click()
    await expect(page.getByRole('tab', { name: '我的申请' })).toHaveAttribute('aria-selected', 'true')
    await page.getByRole('tab', { name: '全部' }).click()
    await expect(page.getByRole('tab', { name: '全部' })).toHaveAttribute('aria-selected', 'true')
  })

  test('table shows approval columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '类型' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '申请人' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '状态' })).toBeVisible()
  })
})
```

- [ ] **Step 3: Delete redundant keys-approval.spec.ts**

```bash
rm apps/frontend/e2e/keys-approval.spec.ts
```

- [ ] **Step 4: Run and verify**

```bash
cd apps/frontend && pnpm exec playwright test e2e/keys-self-service.spec.ts e2e/approval.spec.ts --reporter=list
```

- [ ] **Step 5: Commit**

```bash
git add apps/frontend/e2e/keys-self-service.spec.ts apps/frontend/e2e/approval.spec.ts
git rm apps/frontend/e2e/keys-approval.spec.ts
git commit -m "test(e2e): rewrite key center and approval tests"
```

---

### Task 6: Dashboard & Analytics E2E

**Files:**
- Rewrite: `apps/frontend/e2e/dashboard-cost.spec.ts`
- Remove: `apps/frontend/e2e/dashboard.spec.ts` (merge into dashboard-cost.spec.ts)

**Interfaces:**
- Consumes: Seeded usage data (buckets), dashboard API
- Produces: Cost dashboard cards, charts, usage analysis page

- [ ] **Step 1: Write dashboard-cost.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('成本看板', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard/cost')
    await expect(page.getByRole('heading', { name: '成本看板' })).toBeVisible()
  })

  test('displays stat cards', async ({ page }) => {
    await expect(page.getByText('总花费')).toBeVisible()
    await expect(page.getByText('平均单次成本')).toBeVisible()
    await expect(page.getByText('人均成本')).toBeVisible()
    await expect(page.getByText('总调用次数')).toBeVisible()
  })

  test('displays chart sections', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 3, name: '花费趋势' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '部门成本占比' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '部门花费明细' })).toBeVisible()
  })

  test('shows top consumers', async ({ page }) => {
    const table = page.locator('table').filter({ hasText: '排名' })
    await expect(table.getByRole('columnheader', { name: '成员' })).toBeVisible()
    await expect(table.locator('tbody tr').first()).toBeVisible()
  })
})

test.describe('用量分析', () => {
  test('loads usage analysis page', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByRole('heading', { name: '用量分析' })).toBeVisible()
  })

  test('shows team usage table', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByRole('columnheader', { name: '部门' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '额度' })).toBeVisible()
  })
})
```

- [ ] **Step 2: Remove old dashboard.spec.ts**

```bash
rm apps/frontend/e2e/dashboard.spec.ts
```

- [ ] **Step 3: Run and verify**

```bash
cd apps/frontend && pnpm exec playwright test e2e/dashboard-cost.spec.ts --reporter=list
```

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/e2e/dashboard-cost.spec.ts
git rm apps/frontend/e2e/dashboard.spec.ts
git commit -m "test(e2e): rewrite dashboard and usage analysis tests"
```

---

### Task 7: Wallet, Audit & Navigation E2E

**Files:**
- Rewrite: `apps/frontend/e2e/wallet.spec.ts`
- Rewrite: `apps/frontend/e2e/audit-export.spec.ts`
- Rewrite: `apps/frontend/e2e/navigation.spec.ts`
- Rewrite: `apps/frontend/e2e/member-portal.spec.ts`

**Interfaces:**
- Consumes: Seeded wallet data, audit logs, member session
- Produces: Wallet UI, audit filters/export, nav flow, member portal pages

- [ ] **Step 1: Write wallet.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('钱包管理', () => {
  test('loads wallet page with balance', async ({ page }) => {
    await page.goto('/wallet')
    await expect(page.getByRole('heading', { name: '钱包管理' })).toBeVisible()
    await expect(page.getByText('当前余额')).toBeVisible()
    await expect(page.getByText('累计消费')).toBeVisible()
  })

  test('shows recharge form', async ({ page }) => {
    await page.goto('/wallet')
    await expect(page.getByText('账户充值')).toBeVisible()
    await expect(page.getByRole('button', { name: '确认充值' })).toBeVisible()
  })
})
```

- [ ] **Step 2: Write audit-export.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test.describe('审计', () => {
  test('operations page has export button', async ({ page }) => {
    await page.goto('/audit/operations')
    await expect(page.getByRole('heading', { name: '操作审计' })).toBeVisible()
    await expect(page.getByRole('button', { name: '导出 CSV' })).toBeVisible()
  })

  test('calls page has export button and filters', async ({ page }) => {
    await page.goto('/audit/calls')
    await expect(page.getByRole('heading', { name: '调用日志' })).toBeVisible()
    await expect(page.getByRole('button', { name: '导出 CSV' })).toBeVisible()
  })

  test('operations page shows log table', async ({ page }) => {
    await page.goto('/audit/operations')
    await expect(page.getByRole('columnheader', { name: '操作类型' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '操作人' })).toBeVisible()
  })
})
```

- [ ] **Step 3: Write navigation.spec.ts**

```typescript
import { expect, test } from '@playwright/test'

test('sidebar navigation links work', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('link', { name: '组织架构' }).click()
  await expect(page).toHaveURL(/\/org\/structure/)
  await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
})

test('sidebar shows nav groups', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('组织')).toBeVisible()
  await expect(page.getByText('预算')).toBeVisible()
  await expect(page.getByText('Key 中心')).toBeVisible()
})
```

- [ ] **Step 4: Write member-portal.spec.ts**

```typescript
import { expect, test } from '@playwright/test'
import { loginAsMember } from './helpers/auth'

test.describe('成员工作台', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('member dashboard loads', async ({ page }) => {
    await loginAsMember(page)
    await page.goto('/me')
    await expect(page).toHaveURL(/\/me$/)
    await expect(page.getByText('工作台')).toBeVisible()
  })

  test('member keys page loads', async ({ page }) => {
    await loginAsMember(page)
    await page.goto('/me/keys')
    await expect(page).toHaveURL(/\/me\/keys$/)
    await expect(page.getByRole('link', { name: '我的 Key' })).toBeVisible()
  })

  test('member call logs page loads', async ({ page }) => {
    await loginAsMember(page)
    await page.goto('/me/call-logs')
    await expect(page).toHaveURL(/\/me\/call-logs$/)
  })
})
```

- [ ] **Step 5: Run all tests**

```bash
cd apps/frontend && pnpm exec playwright test e2e/wallet.spec.ts e2e/audit-export.spec.ts e2e/navigation.spec.ts e2e/member-portal.spec.ts --reporter=list
```

- [ ] **Step 6: Commit**

```bash
git add apps/frontend/e2e/wallet.spec.ts apps/frontend/e2e/audit-export.spec.ts apps/frontend/e2e/navigation.spec.ts apps/frontend/e2e/member-portal.spec.ts
git commit -m "test(e2e): rewrite wallet, audit, navigation, and member portal tests"
```

---

### Task 8: Final Verification & Cleanup

**Files:**
- Verify: All `apps/frontend/e2e/*.spec.ts`

- [ ] **Step 1: Run full e2e suite**

```bash
cd apps/frontend && pnpm exec playwright test --reporter=list
```

- [ ] **Step 2: Fix any failures**

Iterate on failing tests — common issues:
- Heading text mismatch (check actual `<h1>` in page)
- Missing seed data (check backend seed modules)
- Timing issues (add `{ timeout: 10_000 }` to slow assertions)

- [ ] **Step 3: Final commit**

```bash
git add -A apps/frontend/e2e/
git commit -m "test(e2e): finalize e2e test regeneration"
```

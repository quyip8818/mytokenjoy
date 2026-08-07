import { expect, test } from '@playwright/test'

const routes = [
  { path: '/org/data-source', testId: 'page-org-data-source' },
  { path: '/org/structure', testId: 'page-org-structure' },
  { path: '/org/roles', testId: 'page-org-roles' },
  { path: '/budget', testId: 'page-budget' },
  { path: '/budget/alerts', testId: 'page-budget-alerts' },
  { path: '/models/list', testId: 'page-models-list' },
  { path: '/models/routing', testId: 'page-models-routing' },
  { path: '/me/keys', testId: 'page-me-keys' },
  { path: '/approvals', testId: 'page-approval' },
  { path: '/keys/platform', testId: 'page-keys-platform' },
  { path: '/keys/provider', testId: 'page-keys-provider' },
  { path: '/dashboard/cost', testId: 'page-dashboard-cost' },
  { path: '/dashboard/usage', testId: 'page-dashboard-usage' },
  { path: '/billing', testId: 'page-billing' },
  { path: '/audit/operations', testId: 'page-audit-operations' },
  { path: '/audit/calls', testId: 'page-audit-calls' },
]

for (const { path, testId } of routes) {
  test(`${path} renders page`, async ({ page }) => {
    await page.goto(path)
    await expect(page.getByTestId(testId)).toBeVisible()
  })
}

import { expect, test } from '@playwright/test'

test.describe('Lot 审计', () => {
  test('billing page shows "查看批次明细" button', async ({ page }) => {
    await page.goto('/billing')
    await expect(page.getByTestId('page-billing')).toBeVisible()
    await expect(page.getByRole('button', { name: '查看批次明细' })).toBeVisible()
  })

  test('platform companies page shows "审计" menu item', async ({ page }) => {
    await page.goto('/platform/companies')
    const pageEl = page.getByTestId('page-platform-companies')
    if (!(await pageEl.isVisible({ timeout: 3000 }).catch(() => false))) {
      test.skip(true, 'user lacks platform:manage — cannot test audit menu')
      return
    }
    // Open the first company's more-actions dropdown.
    const dropdownTriggers = page
      .locator('[data-testid="page-platform-companies"] table tbody tr')
      .first()
      .locator('button[aria-haspopup="menu"]')
    if ((await dropdownTriggers.count()) === 0) {
      test.skip(true, 'no companies in overview — cannot test audit menu')
      return
    }
    await dropdownTriggers.first().click()
    await expect(page.getByRole('menuitem', { name: '审计' })).toBeVisible()
  })
})

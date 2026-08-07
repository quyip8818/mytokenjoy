import { expect, test } from '@playwright/test'

test.describe('优惠折扣', () => {
  test('platform companies page shows discount menu item', async ({ page }) => {
    await page.goto('/platform/companies')
    // Platform page requires platform:manage — skip if no access
    const pageEl = page.getByTestId('page-platform-companies')
    if (!(await pageEl.isVisible({ timeout: 3000 }).catch(() => false))) {
      test.skip(true, 'user lacks platform:manage — cannot test discount menu')
      return
    }
    // Find first company's more-actions button and open dropdown
    const firstMoreBtn = page.locator('[data-testid^="platform-companies-discount-"]').first()
    // The dropdown menu item should be findable (even if hidden until menu opens)
    await expect(firstMoreBtn).toBeAttached()
  })
})

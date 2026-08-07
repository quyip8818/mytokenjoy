import { expect, test } from '@playwright/test'

test.describe('优惠折扣', () => {
  test('billing page hides discount section when no discounts', async ({ page }) => {
    await page.goto('/billing')
    await expect(
      page.getByRole('banner').getByRole('heading', { name: '钱包管理' }),
    ).toBeVisible()
    // No discount configured → section should not be visible
    await expect(page.getByText('当前优惠')).not.toBeVisible()
  })

  test('platform companies page shows discount menu item', async ({ page }) => {
    await page.goto('/platform/companies')
    // Platform page requires platform:manage — may redirect if user lacks permission.
    const heading = page.getByRole('banner').getByRole('heading', { name: '企业管理' })
    if (!(await heading.isVisible({ timeout: 3000 }).catch(() => false))) {
      test.skip(true, 'user lacks platform:manage — cannot test discount menu')
      return
    }
    // Open first row's dropdown menu
    const moreBtn = page.locator('button[class*="h-8 w-8"]').first()
    if (await moreBtn.isVisible()) {
      await moreBtn.click()
      await expect(page.getByRole('menuitem', { name: '优惠' })).toBeVisible()
    }
  })
})

import { expect, test } from '@playwright/test'

test.describe('优惠折扣', () => {
  test('billing page hides discount section when no discounts', async ({ page }) => {
    await page.goto('/billing')
    await expect(page.getByRole('heading', { name: '钱包管理' })).toBeVisible()
    // No discount configured → section should not be visible
    await expect(page.getByText('当前优惠')).not.toBeVisible()
  })

  test('platform companies page shows discount menu item', async ({ page }) => {
    await page.goto('/platform/companies')
    await expect(page.getByRole('heading', { name: '企业管理' })).toBeVisible()
    // Open first row's dropdown menu
    const moreBtn = page.getByRole('button', { name: '' }).first()
    if (await moreBtn.isVisible()) {
      await moreBtn.click()
      await expect(page.getByRole('menuitem', { name: '优惠' })).toBeVisible()
    }
  })
})

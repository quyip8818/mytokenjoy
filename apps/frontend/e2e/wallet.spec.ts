import { expect, test } from '@playwright/test'

test.describe('钱包管理', () => {
  test('loads wallet page with balance', async ({ page }) => {
    await page.goto('/billing')
    await expect(page.getByTestId('page-billing')).toBeVisible()
    await expect(page.getByText('当前余额')).toBeVisible()
  })

  test('shows recharge section', async ({ page }) => {
    await page.goto('/billing')
    await expect(page.getByTestId('page-billing')).toBeVisible()
    // Recharge panel uses heading or button — use text assertion for recharge presence
    await expect(page.getByRole('button', { name: '充值' })).toBeVisible()
  })
})

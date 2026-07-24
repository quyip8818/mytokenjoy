import { expect, test } from '@playwright/test'

test.describe('钱包管理', () => {
  test('loads wallet page with balance', async ({ page }) => {
    await page.goto('/billing')
    await expect(page.getByRole('banner').getByRole('heading', { name: '钱包管理' })).toBeVisible()
    await expect(page.getByText('当前余额')).toBeVisible()
  })

  test('shows recharge section', async ({ page }) => {
    await page.goto('/billing')
    await expect(page.getByRole('heading', { name: '充值开票' })).toBeVisible()
    await expect(page.getByRole('button', { name: '充值记录' })).toBeVisible()
  })
})

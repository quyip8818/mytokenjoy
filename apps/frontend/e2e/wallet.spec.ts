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

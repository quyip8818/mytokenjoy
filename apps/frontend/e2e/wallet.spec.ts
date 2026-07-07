import { expect, test } from '@playwright/test'

test('loads wallet page', async ({ page }) => {
  await page.goto('/wallet')

  await expect(page).toHaveURL(/\/wallet$/)
  await expect(page.getByRole('main').getByRole('heading', { name: '钱包管理' })).toBeVisible()
  await expect(page.getByText('账户充值')).toBeVisible()
  await expect(page.getByRole('button', { name: '确认充值' })).toBeVisible()
})

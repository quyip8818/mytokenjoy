import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads wallet page with recharge section', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/wallet')

  await expect(page).toHaveURL(/\/wallet$/)
  await expect(page.getByRole('heading', { name: '钱包管理' })).toBeVisible()
  await expect(page.getByText('账户充值')).toBeVisible()
  await expect(page.getByRole('button', { name: '确认充值' })).toBeVisible()
})

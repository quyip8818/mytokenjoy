import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads budget page with department tree', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/budget')

  await expect(page).toHaveURL(/\/budget$/)
  await expect(page.getByText('选择左侧节点查看预算详情')).toBeVisible()
})

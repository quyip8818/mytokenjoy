import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads data source platform selection', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/org/data-source')

  await expect(page).toHaveURL(/\/org\/data-source$/)
  await expect(page.getByRole('button', { name: /飞书/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /钉钉/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /企业微信/ })).toBeVisible()
})

import { expect, test } from '@playwright/test'
import { loginAsMember } from './helpers/auth'

test('member portal dashboard and keys pages load', async ({ page }) => {
  await loginAsMember(page)

  await page.goto('/me')
  await expect(page).toHaveURL(/\/me$/)
  await expect(page.getByText('工作台')).toBeVisible()

  await page.goto('/me/keys')
  await expect(page).toHaveURL(/\/me\/keys$/)
  await expect(page.getByRole('link', { name: '我的 Key' })).toBeVisible()
})

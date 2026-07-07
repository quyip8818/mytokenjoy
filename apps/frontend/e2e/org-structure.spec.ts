import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads org structure member list', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/org/structure')

  await expect(page).toHaveURL(/\/org\/structure$/)
  await expect(page.getByRole('columnheader', { name: '姓名' })).toBeVisible()
})

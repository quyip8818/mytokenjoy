import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads org roles member list after selecting a role', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/org/roles')

  await expect(page).toHaveURL(/\/org\/roles$/)
  await page.getByText('超级管理员').first().click()
  await expect(page.getByText(/名成员/)).toBeVisible()
})

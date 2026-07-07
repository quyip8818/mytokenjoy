import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads keys approval page', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/keys/approval')

  await expect(page).toHaveURL(/\/keys\/approval$/)
  await expect(page.getByText('待我审批')).toBeVisible()
})

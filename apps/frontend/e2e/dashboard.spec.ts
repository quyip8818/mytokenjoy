import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads cost dashboard for authenticated admin', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/dashboard/cost')

  await expect(page).toHaveURL(/\/dashboard\/cost$/)
  await expect(page.getByRole('main')).toBeVisible()
  await expect(page.getByText('总花费')).toBeVisible()
})

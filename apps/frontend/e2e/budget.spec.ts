import { expect, test } from '@playwright/test'

test('loads budget page with department tree', async ({ page }) => {
  await page.goto('/budget')

  await expect(page).toHaveURL(/\/budget$/)
  await expect(page.getByRole('button', { name: '上一月' })).toBeVisible()
})

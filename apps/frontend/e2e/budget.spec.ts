import { expect, test } from '@playwright/test'

test('loads budget page with department tree', async ({ page }) => {
  await page.goto('/budget')

  await expect(page).toHaveURL(/\/budget$/)
  await expect(page.getByRole('treeitem', { name: '总公司' })).toBeVisible()
  await expect(page.getByRole('heading', { level: 3, name: '总公司' })).toBeVisible()
})

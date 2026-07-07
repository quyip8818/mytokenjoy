import { expect, test } from '@playwright/test'

test('sidebar navigation links work', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('link', { name: '组织架构' }).click()
  await expect(page).toHaveURL(/\/org\/structure/)
  await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
})

test('sidebar shows nav groups', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('组织')).toBeVisible()
  await expect(page.getByText('预算')).toBeVisible()
  await expect(page.getByText('Key 中心')).toBeVisible()
})

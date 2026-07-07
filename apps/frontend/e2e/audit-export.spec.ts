import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('loads audit pages and export button is clickable', async ({ page }) => {
  await loginAsAdmin(page)

  await page.goto('/audit/operations')
  await expect(page).toHaveURL(/\/audit\/operations$/)
  await expect(page.getByRole('button', { name: '导出 CSV' })).toBeVisible()
  await page.getByRole('button', { name: '导出 CSV' }).click()

  await page.goto('/audit/calls')
  await expect(page).toHaveURL(/\/audit\/calls$/)
  await expect(page.getByRole('button', { name: '导出 CSV' })).toBeVisible()
  await page.getByRole('button', { name: '导出 CSV' }).click()
})

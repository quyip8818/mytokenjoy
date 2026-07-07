import { expect, test } from '@playwright/test'

test.describe('审计', () => {
  test('operations page has export button', async ({ page }) => {
    await page.goto('/audit/operations')
    await expect(page.getByRole('heading', { name: '操作审计' })).toBeVisible()
    await expect(page.getByRole('button', { name: '导出 CSV' })).toBeVisible()
  })

  test('calls page has export button and filters', async ({ page }) => {
    await page.goto('/audit/calls')
    await expect(page.getByRole('heading', { name: '调用日志' })).toBeVisible()
    await expect(page.getByRole('button', { name: '导出 CSV' })).toBeVisible()
  })

  test('operations page shows log table', async ({ page }) => {
    await page.goto('/audit/operations')
    await expect(page.getByRole('columnheader', { name: '操作类型' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '操作人' })).toBeVisible()
  })
})

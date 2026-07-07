import { expect, test } from '@playwright/test'

test.describe('成本看板', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard/cost')
    await expect(page.getByRole('banner').getByRole('heading', { name: '成本看板' })).toBeVisible()
  })

  test('displays summary stat cards', async ({ page }) => {
    await expect(page.getByText('总花费')).toBeVisible()
    await expect(page.getByText('平均单次成本')).toBeVisible()
    await expect(page.getByText('人均成本')).toBeVisible()
    await expect(page.getByText('总调用次数')).toBeVisible()
    await expect(page.getByText('总 Token')).toBeVisible()
  })

  test('displays charts and tables', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 3, name: '花费趋势' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '部门成本占比' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '部门花费明细' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '消费排行 Top 5' })).toBeVisible()
  })

  test('shows top consumers in ranking table', async ({ page }) => {
    const table = page.locator('table').filter({ hasText: '排名' })
    await expect(table.getByRole('columnheader', { name: '成员' })).toBeVisible()
    await expect(table.getByRole('columnheader', { name: '花费' })).toBeVisible()
    await expect(table.getByRole('columnheader', { name: '请求数' })).toBeVisible()
    // At least one row of data
    await expect(table.locator('tbody tr').first()).toBeVisible()
  })
})

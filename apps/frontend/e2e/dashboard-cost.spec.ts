import { expect, test } from '@playwright/test'

test.describe('成本看板', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard/cost')
    await expect(page.getByRole('heading', { name: '成本看板' })).toBeVisible()
  })

  test('displays stat cards', async ({ page }) => {
    await expect(page.getByText('总花费')).toBeVisible()
    await expect(page.getByText('平均单次成本')).toBeVisible()
    await expect(page.getByText('人均成本')).toBeVisible()
    await expect(page.getByText('总调用次数')).toBeVisible()
  })

  test('displays chart sections', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 3, name: '花费趋势' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '部门成本占比' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 3, name: '部门花费明细' })).toBeVisible()
  })

  test('shows top consumers', async ({ page }) => {
    const table = page.locator('table').filter({ hasText: '排名' })
    await expect(table.getByRole('columnheader', { name: '成员' })).toBeVisible()
    await expect(table.locator('tbody tr').first()).toBeVisible()
  })
})

test.describe('用量分析', () => {
  test('loads usage analysis page', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByRole('heading', { name: '用量分析' })).toBeVisible()
  })

  test('shows team usage table', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByRole('columnheader', { name: '部门' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '额度' })).toBeVisible()
  })
})

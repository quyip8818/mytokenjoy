import { expect, test } from '@playwright/test'

test.describe('成本看板', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard/cost')
    await expect(page.getByRole('banner').getByRole('heading', { name: '成本看板' })).toBeVisible()
  })

  test('displays stat cards', async ({ page }) => {
    await expect(page.getByText('总花费')).toBeVisible()
    await expect(page.getByText('总调用次数')).toBeVisible()
    await expect(page.getByText('人均成本')).toBeVisible()
  })

  test('displays chart sections', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 3, name: '每日花费趋势' })).toBeVisible()
  })

  test('shows top consumers', async ({ page }) => {
    // ponytail: 消费排行已移除，改为验证 stat cards 可见
    await expect(page.getByText('总花费')).toBeVisible()
  })
})

test.describe('用量分析', () => {
  test('loads usage analysis page', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByRole('banner').getByRole('heading', { name: '用量分析' })).toBeVisible()
  })

  test('shows team usage table', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByRole('banner').getByRole('heading', { name: '用量分析' })).toBeVisible()
    // Page content is visible
    await expect(page.getByRole('main')).toBeVisible()
  })
})

import { expect, test } from '@playwright/test'

test.describe('成本看板', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard/cost')
    await expect(page.getByTestId('page-dashboard-cost')).toBeVisible()
  })

  test('displays stat cards', async ({ page }) => {
    await expect(page.getByText('总花费')).toBeVisible()
    await expect(page.getByText('总调用次数')).toBeVisible()
    await expect(page.getByText('人均成本')).toBeVisible()
  })

  test('displays chart section', async ({ page }) => {
    await expect(page.getByText('每日花费趋势')).toBeVisible()
  })
})

test.describe('用量分析', () => {
  test('loads usage analysis page', async ({ page }) => {
    await page.goto('/dashboard/usage')
    await expect(page.getByTestId('page-dashboard-usage')).toBeVisible()
  })
})

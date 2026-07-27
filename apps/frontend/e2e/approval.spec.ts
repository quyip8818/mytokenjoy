import { expect, test } from '@playwright/test'

test.describe('审批中心', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/approvals')
    await expect(page.getByRole('banner').getByRole('heading', { name: '审批中心' })).toBeVisible()
  })

  test('displays approval tabs', async ({ page }) => {
    await expect(page.getByRole('tab', { name: '待审批' })).toBeVisible()
    await expect(page.getByRole('tab', { name: '已通过' })).toBeVisible()
    await expect(page.getByRole('tab', { name: '已拒绝' })).toBeVisible()
    await expect(page.getByRole('tab', { name: '全部' })).toBeVisible()
  })

  test('default tab is 待审批', async ({ page }) => {
    await expect(page.getByRole('tab', { name: '待审批' })).toHaveAttribute('aria-selected', 'true')
  })

  test('switches between tabs', async ({ page }) => {
    await page.getByRole('tab', { name: '已通过' }).click()
    await expect(page.getByRole('tab', { name: '已通过' })).toHaveAttribute('aria-selected', 'true')
    await page.getByRole('tab', { name: '全部' }).click()
    await expect(page.getByRole('tab', { name: '全部' })).toHaveAttribute('aria-selected', 'true')
  })

  test('shows approval list heading', async ({ page }) => {
    await expect(page.getByRole('main').getByRole('heading', { name: '审批中心' })).toBeVisible()
  })
})

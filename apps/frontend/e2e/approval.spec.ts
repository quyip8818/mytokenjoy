import { expect, test } from '@playwright/test'

test.describe('审批中心', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/approvals')
    await expect(page.getByRole('heading', { name: '审批中心' })).toBeVisible()
  })

  test('displays approval tabs', async ({ page }) => {
    await expect(page.getByRole('tab', { name: /待我审批/ })).toBeVisible()
    await expect(page.getByRole('tab', { name: '我的申请' })).toBeVisible()
    await expect(page.getByRole('tab', { name: '全部' })).toBeVisible()
  })

  test('default tab is 待我审批', async ({ page }) => {
    await expect(page.getByRole('tab', { name: /待我审批/ })).toHaveAttribute(
      'aria-selected',
      'true',
    )
  })

  test('switches between tabs', async ({ page }) => {
    await page.getByRole('tab', { name: '我的申请' }).click()
    await expect(page.getByRole('tab', { name: '我的申请' })).toHaveAttribute(
      'aria-selected',
      'true',
    )
    await page.getByRole('tab', { name: '全部' }).click()
    await expect(page.getByRole('tab', { name: '全部' })).toHaveAttribute('aria-selected', 'true')
  })

  test('table shows approval columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '类型' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '申请人' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '状态' })).toBeVisible()
  })
})

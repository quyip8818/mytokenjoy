import { expect, test } from '@playwright/test'

test.describe('预算管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/budget')
    await expect(page.getByTestId('page-budget')).toBeVisible()
  })

  test('displays budget tree', async ({ page }) => {
    await expect(page.getByRole('treeitem').first()).toBeVisible()
  })

  test('selecting a node shows detail panel', async ({ page }) => {
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    // Either shows budget info or "设置总额度" prompt
    await expect(
      page.getByText(/已分配|总预算|已使用|尚未设置预算额度|设置总额度/).first(),
    ).toBeVisible()
  })
})

test.describe('预警规则', () => {
  test('loads alerts page with rule list', async ({ page }) => {
    await page.goto('/budget/alerts')
    await expect(page.getByTestId('page-budget-alerts')).toBeVisible()
    await expect(page.getByRole('button', { name: /新建规则|添加|创建/ })).toBeVisible()
  })
})

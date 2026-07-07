import { expect, test } from '@playwright/test'

test.describe('预算管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/budget')
    await expect(page.getByRole('heading', { name: '预算管理' })).toBeVisible()
  })

  test('displays month navigation', async ({ page }) => {
    await expect(page.getByRole('button', { name: '上一月' })).toBeVisible()
    await expect(page.getByRole('button', { name: '下一月' })).toBeVisible()
  })

  test('displays budget tree', async ({ page }) => {
    await expect(page.getByRole('treeitem').first()).toBeVisible()
  })

  test('selecting a node shows detail panel', async ({ page }) => {
    await page.getByRole('treeitem').first().click()
    await expect(page.getByText(/已分配|总预算|已使用/)).toBeVisible()
  })
})

test.describe('预警规则', () => {
  test('loads alerts page with rule list', async ({ page }) => {
    await page.goto('/budget/alerts')
    await expect(page.getByRole('heading', { name: '预警规则' })).toBeVisible()
    await expect(page.getByRole('button', { name: /新建规则|添加/ })).toBeVisible()
  })
})

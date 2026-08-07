import { expect, test } from '@playwright/test'

test('sidebar navigation links work', async ({ page }) => {
  await page.goto('/')
  // Expand the '组织与权限' group
  await page.getByTestId('nav-group-组织与权限').click()
  // Click org structure nav item
  await page.getByTestId('nav-org-structure').click()
  await expect(page).toHaveURL(/\/org\/structure/)
  await expect(page.getByTestId('page-org-structure')).toBeVisible()
})

test('sidebar shows nav groups', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('nav-group-数据看板')).toBeVisible()
  await expect(page.getByTestId('nav-group-凭证管理')).toBeVisible()
  await expect(page.getByTestId('nav-group-模型管理')).toBeVisible()
  await expect(page.getByTestId('nav-group-预算与财务')).toBeVisible()
  await expect(page.getByTestId('nav-group-组织与权限')).toBeVisible()
})

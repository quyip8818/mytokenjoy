import { expect, test } from '@playwright/test'

test('sidebar navigation links work', async ({ page }) => {
  await page.goto('/')
  // Expand the '组织与权限' group first
  await page.getByRole('button', { name: '组织与权限' }).click()
  await page.getByRole('link', { name: '组织架构' }).click()
  await expect(page).toHaveURL(/\/org\/structure/)
  await expect(page.getByRole('banner').getByRole('heading', { name: '组织架构' })).toBeVisible()
})

test('sidebar shows nav groups', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('button', { name: '数据看板' })).toBeVisible()
  await expect(page.getByRole('button', { name: '凭证管理' })).toBeVisible()
  await expect(page.getByRole('button', { name: '模型管理' })).toBeVisible()
  await expect(page.getByRole('button', { name: '预算与财务' })).toBeVisible()
  await expect(page.getByRole('button', { name: '组织与权限' })).toBeVisible()
})

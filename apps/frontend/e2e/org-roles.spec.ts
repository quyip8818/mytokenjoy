import { expect, test } from '@playwright/test'

test.describe('角色管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('displays preset roles', async ({ page }) => {
    await expect(page.getByText('超级管理员')).toBeVisible()
    await expect(page.getByText('普通成员')).toBeVisible()
  })

  test('selecting a role shows member list', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await expect(page.getByText(/名成员/)).toBeVisible()
  })
})

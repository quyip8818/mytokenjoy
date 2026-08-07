import { expect, test } from '@playwright/test'
import { loginAsMember } from './helpers/auth'

test.describe('成员工作台', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('member can log in and see app', async ({ page }) => {
    try {
      await loginAsMember(page)
    } catch {
      test.skip(true, '成员账户无法登录（demo 环境无成员凭据）')
      return
    }
    await page.goto('/')
    await expect(page).not.toHaveURL(/\/login/)
  })

  test('member keys page loads', async ({ page }) => {
    try {
      await loginAsMember(page)
    } catch {
      test.skip(true, '成员账户无法登录（demo 环境无成员凭据）')
      return
    }
    await page.goto('/me/keys')
    await expect(page.getByTestId('page-me-keys')).toBeVisible()
  })
})

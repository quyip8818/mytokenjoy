import { expect, test } from '@playwright/test'
import { loginAsMember } from './helpers/auth'

test.describe('成员工作台', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('member dashboard loads', async ({ page }) => {
    await loginAsMember(page)
    await page.goto('/me')
    await expect(page).toHaveURL(/\/me$/)
    await expect(page.getByText('工作台')).toBeVisible()
  })

  test('member keys page loads', async ({ page }) => {
    await loginAsMember(page)
    await page.goto('/me/keys')
    await expect(page).toHaveURL(/\/me\/keys$/)
    await expect(page.getByRole('link', { name: '我的 Key' })).toBeVisible()
  })

  test('member call logs page loads', async ({ page }) => {
    await loginAsMember(page)
    await page.goto('/me/call-logs')
    await expect(page).toHaveURL(/\/me\/call-logs$/)
  })
})

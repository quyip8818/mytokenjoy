import { expect, test } from '@playwright/test'

// Tests without auth state
test.use({ storageState: { cookies: [], origins: [] } })

test('redirects unauthenticated user to /login', async ({ page }) => {
  await page.goto('/org/structure')
  await expect(page).toHaveURL(/\/login/)
})

test('renders login form fields', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByLabel('邮箱')).toBeVisible()
  await expect(page.getByLabel('密码')).toBeVisible()
  await expect(page.getByRole('button', { name: '登录' }).first()).toBeVisible()
})

test('login with valid credentials redirects to app', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByLabel('邮箱')).toBeVisible()
  await page.getByLabel('邮箱').fill('demo@tokenjoy.me')
  await page.getByLabel('密码').fill('demo1234')
  await page.locator('form button[type="submit"]').click()
  await expect(page).not.toHaveURL(/\/login/, { timeout: 10_000 })
})

test('login with invalid credentials shows error', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('邮箱').fill('demo@tokenjoy.me')
  await page.getByLabel('密码').fill('wrongpass')
  await page.locator('form button[type="submit"]').click()
  await expect(page).toHaveURL(/\/login/)
})

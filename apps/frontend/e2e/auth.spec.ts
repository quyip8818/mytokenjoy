import { expect, test } from '@playwright/test'

// Tests without auth state
test.use({ storageState: { cookies: [], origins: [] } })

test('redirects unauthenticated user to /login', async ({ page }) => {
  await page.goto('/org/structure')
  await expect(page).toHaveURL(/\/login/)
})

test('renders login form fields', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByLabel('Email')).toBeVisible()
  await expect(page.getByLabel('Password')).toBeVisible()
  await expect(page.getByRole('button', { name: '登录' })).toBeVisible()
})

test('login with valid credentials redirects to app', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill('admin@example.com')
  await page.getByLabel('Password').fill('demo1234')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page).not.toHaveURL(/\/login/)
})

test('login with invalid credentials shows error', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill('admin@example.com')
  await page.getByLabel('Password').fill('wrongpass')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page).toHaveURL(/\/login/)
})

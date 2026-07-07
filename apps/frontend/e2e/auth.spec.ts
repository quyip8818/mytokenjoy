import { expect, test } from '@playwright/test'

// This test must run without stored auth state
test.use({ storageState: { cookies: [], origins: [] } })

test('redirects unauthenticated users to login', async ({ page }) => {
  await page.goto('/org/structure')
  await expect(page).toHaveURL(/\/login/)
})

test('renders login form on login page', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByLabel('Email')).toBeVisible()
  await expect(page.getByLabel('Password')).toBeVisible()
})

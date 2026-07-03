import { expect, test } from '@playwright/test'

test('redirects unauthenticated users to login', async ({ page }) => {
  await page.goto('/org/structure')
  await expect(page).toHaveURL(/\/login$/)
  await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible()
})

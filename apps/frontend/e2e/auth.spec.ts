import { expect, test } from '@playwright/test'

// This test must run without stored auth state
test.use({ storageState: { cookies: [], origins: [] } })

test('redirects unauthenticated users to login', async ({ page }) => {
  const response = await page.goto('/org/structure')
  // Verify we end up on the login page
  await expect(page).toHaveURL(/\/login/)
  // The redirect itself is the key assertion — login page rendering
  // is verified by the full auth flow in other tests
})

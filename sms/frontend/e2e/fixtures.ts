import { test as base, type Page } from '@playwright/test'

async function login(page: Page, username = 'admin', password = 'admin123') {
  await page.goto('/login')
  await page.getByRole('textbox').first().fill(username)
  await page.getByRole('textbox').nth(1).fill(password)
  await page.getByRole('button', { name: /登录/ }).click()
  await page.waitForURL('**/dashboard**')
}

export const test = base.extend<{ authedPage: Page }>({
  authedPage: async ({ page }, use) => {
    await login(page)
    await use(page)
  },
})

export { expect } from '@playwright/test'

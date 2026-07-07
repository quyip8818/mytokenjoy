import { chromium } from '@playwright/test'

const BASE_URL = 'http://127.0.0.1:4173'

async function loginAndSave(email: string, password: string, savePath: string) {
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  // Use API login (faster than UI)
  const response = await page.request.post(`${BASE_URL}/api/auth/login`, {
    data: { email, password },
  })
  if (!response.ok()) {
    throw new Error(`Login failed for ${email}: ${response.status()}`)
  }

  // Navigate to trigger cookie attachment to browser context
  await page.goto(BASE_URL)

  await context.storageState({ path: savePath })
  await browser.close()
}

export default async function globalSetup() {
  await loginAndSave('admin@example.com', 'demo1234', '.auth/admin.json')
}

import { chromium } from '@playwright/test'
import { E2E_BASE_URL } from './e2e-db'

async function loginAndSave(email: string, password: string, savePath: string) {
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  const response = await page.request.post(`${E2E_BASE_URL}/api/auth/login`, {
    data: { email, password },
  })
  if (!response.ok()) {
    throw new Error(`Login failed for ${email}: ${response.status()}`)
  }

  await page.goto(E2E_BASE_URL)
  await context.storageState({ path: savePath })
  await browser.close()
}

export default async function globalSetup() {
  // Both modes use BootstrapDemo with same seed data — demo admin works in both.
  await loginAndSave('demo@tokenjoy.me', 'demo1234', '.auth/admin.json')
}

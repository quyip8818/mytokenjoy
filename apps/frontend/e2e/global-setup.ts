import { chromium } from '@playwright/test'
import { E2E_BASE_URL, E2E_MODE } from './e2e-db'

const credentials = {
  saas: { email: 'admin@tokenjoy.me', password: 'admin1234' },
  local: { email: 'demo@tokenjoy.me', password: 'demo1234' },
}

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
  const { email, password } = credentials[E2E_MODE]
  await loginAndSave(email, password, '.auth/admin.json')
}

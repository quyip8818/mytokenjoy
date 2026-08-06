import { chromium } from '@playwright/test'
import { E2E_BASE_URL } from './e2e-db'

async function loginAndSave(email: string, password: string, savePath: string) {
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  // Step 1: Authenticate — gets register cookie + company list
  const loginResp = await page.request.post(`${E2E_BASE_URL}/api/auth/login`, {
    data: { email, password },
  })
  if (!loginResp.ok()) {
    throw new Error(`Login failed for ${email}: ${loginResp.status()}`)
  }
  const loginData = await loginResp.json()

  // Step 2: Select company — gets session cookie
  if (loginData.action === 'select_company' && loginData.companies?.length > 0) {
    const selectResp = await page.request.post(`${E2E_BASE_URL}/api/auth/select-company`, {
      data: { companyId: loginData.companies[0].companyId },
    })
    if (!selectResp.ok()) {
      throw new Error(`Select company failed: ${selectResp.status()}`)
    }
  }

  await page.goto(E2E_BASE_URL)
  await context.storageState({ path: savePath })
  await browser.close()
}

export default async function globalSetup() {
  // Both modes use BootstrapDemo with same seed data — demo admin works in both.
  await loginAndSave('demo@tokenjoy.me', 'demo1234', '.auth/admin.json')
}

/**
 * Playwright login helper — can be run standalone or imported.
 *
 * Standalone: npx tsx e2e/login.ts [baseUrl]
 * Logs in as admin and saves storageState to .auth/admin.json
 */

const DEFAULT_BASE_URL = 'http://127.0.0.1:5173'
const CREDENTIALS = { email: 'admin@example.com', password: 'demo1234' }
const STORAGE_PATH = '.auth/admin.json'

export async function login(baseUrl = DEFAULT_BASE_URL) {
  const { chromium } = await import('@playwright/test')
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  const response = await page.request.post(`${baseUrl}/api/auth/login`, {
    data: CREDENTIALS,
  })
  if (!response.ok()) {
    throw new Error(`Login failed: ${response.status()} ${response.statusText()}`)
  }

  await page.goto(baseUrl)
  await context.storageState({ path: STORAGE_PATH })
  await browser.close()
  console.log(`✓ Logged in as ${CREDENTIALS.email}, saved to ${STORAGE_PATH}`)
}

// Run standalone
if (process.argv[1]?.endsWith('login.ts')) {
  const baseUrl = process.argv[2] || DEFAULT_BASE_URL
  login(baseUrl).catch((err) => {
    console.error(err.message)
    process.exit(1)
  })
}

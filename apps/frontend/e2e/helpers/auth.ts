import type { APIRequestContext, Page } from '@playwright/test'

const DEMO_EMAIL = 'admin@example.com'
const DEMO_PASSWORD = 'demo1234'

async function postLogin(request: APIRequestContext, baseURL: string): Promise<void> {
  const response = await request.post(`${baseURL}/api/auth/login`, {
    data: { email: DEMO_EMAIL, password: DEMO_PASSWORD },
  })
  if (!response.ok()) {
    throw new Error(`login failed: ${response.status()} ${await response.text()}`)
  }
}

export async function loginAsAdmin(page: Page): Promise<void> {
  const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4173'
  await postLogin(page.request, baseURL)
}

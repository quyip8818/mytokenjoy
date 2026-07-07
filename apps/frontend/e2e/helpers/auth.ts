import type { APIRequestContext, Page } from '@playwright/test'

const DEMO_EMAIL = 'admin@example.com'
const DEMO_PASSWORD = 'demo1234'
const MEMBER_EMAIL = 'zhangsan@example.com'

async function postLogin(
  request: APIRequestContext,
  baseURL: string,
  email: string,
  password: string,
): Promise<void> {
  const response = await request.post(`${baseURL}/api/auth/login`, {
    data: { email, password },
  })
  if (!response.ok()) {
    throw new Error(`login failed: ${response.status()} ${await response.text()}`)
  }
}

export async function loginAsAdmin(page: Page): Promise<void> {
  const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4173'
  await postLogin(page.request, baseURL, DEMO_EMAIL, DEMO_PASSWORD)
}

export async function loginAsMember(page: Page): Promise<void> {
  const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4173'
  await postLogin(page.request, baseURL, MEMBER_EMAIL, DEMO_PASSWORD)
}

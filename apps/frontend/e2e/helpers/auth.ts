import type { APIRequestContext, Page } from '@playwright/test'
import { E2E_BASE_URL } from '../e2e-db'

const DEMO_EMAIL = 'demo@tokenjoy.me'
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
  const body = await response.json()
  if (body.action) {
    throw new Error(`login incomplete: action=${body.action}`)
  }
}

export async function loginAsAdmin(page: Page): Promise<void> {
  await postLogin(page.request, E2E_BASE_URL, DEMO_EMAIL, DEMO_PASSWORD)
}

export async function loginAsMember(page: Page): Promise<void> {
  await postLogin(page.request, E2E_BASE_URL, MEMBER_EMAIL, DEMO_PASSWORD)
}

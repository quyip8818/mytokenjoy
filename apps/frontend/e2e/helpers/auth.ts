import type { Page } from '@playwright/test'

export const SESSION_COOKIE = 'tokenjoy_session_member'
export const ADMIN_MEMBER_ID = 'm-admin'

export async function loginAsAdmin(page: Page): Promise<void> {
  await page.context().addCookies([
    {
      name: SESSION_COOKIE,
      value: ADMIN_MEMBER_ID,
      domain: '127.0.0.1',
      path: '/',
      sameSite: 'Lax',
    },
  ])
}

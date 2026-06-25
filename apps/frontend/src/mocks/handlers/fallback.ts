import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'

export function isMockApiUrl(url: string): boolean {
  const { pathname } = new URL(url)
  return pathname === API_BASE_PATH || pathname.startsWith(`${API_BASE_PATH}/`)
}

export const fallbackHandlers = [
  http.all(
    ({ request }) => isMockApiUrl(request.url),
    ({ request }) => {
      const { pathname } = new URL(request.url)
      return HttpResponse.json(
        { message: `No mock handler for ${request.method} ${pathname}` },
        { status: 501 },
      )
    },
  ),
]

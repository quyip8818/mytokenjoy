import { API_BASE_PATH } from '@/config/app'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

let demoMemberIdProvider: (() => string | null) | null = null

export function setDemoMemberIdProvider(provider: () => string | null): void {
  demoMemberIdProvider = provider
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_PATH}${path}`
  const memberId = demoMemberIdProvider?.()
  const res = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...(memberId ? { 'X-Demo-Member-Id': memberId } : {}),
      ...options.headers,
    },
    ...options,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body.message || res.statusText)
  }

  return res.json()
}

export function buildQuery(params: object): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    search.set(key, String(value))
  }
  const qs = search.toString()
  return qs ? `?${qs}` : ''
}

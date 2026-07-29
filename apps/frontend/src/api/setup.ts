import { API_BASE_PATH } from '@/config/app'

export interface SetupStatus {
  ready: boolean
}

export interface SetupInitInput {
  companyName: string
  industry: string
  size: string
  adminEmail: string
  adminPassword: string
  adminName: string
}

export interface SetupInitResult {
  companyId: string
  status: string
}

// ponytail: setup API 不走通用 request()（那个依赖 session），直接 fetch。
// Setup server 运行在 bootstrap 之前，没有 auth 中间件。

async function setupRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_PATH}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || `Setup request failed: ${res.status}`)
  }
  return res.json()
}

export const setupApi = {
  status: () => setupRequest<SetupStatus>('/setup/status'),

  init: (input: SetupInitInput) =>
    setupRequest<SetupInitResult>('/setup/init', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
}

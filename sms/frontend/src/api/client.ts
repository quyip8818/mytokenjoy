import { API_BASE_PATH } from '@/config/app'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

let refreshing: Promise<boolean> | null = null

function doRefresh(): Promise<boolean> {
  if (!refreshing) {
    refreshing = fetch(`${API_BASE_PATH}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
      .then(async (r) => {
        if (!r.ok) return false
        const data = await r.json()
        localStorage.setItem('sms_access_token', data.accessToken)
        return true
      })
      .catch(() => false)
      .finally(() => {
        refreshing = null
      })
  }
  return refreshing
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_PATH}${path}`
  const token = localStorage.getItem('sms_access_token')
  const headers: Record<string, string> = {
    Accept: 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  }
  // Don't set Content-Type for FormData (let browser set boundary)
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  const init: RequestInit = {
    credentials: 'include',
    ...options,
    headers: { ...headers, ...(options.headers as Record<string, string>) },
  }

  let res = await fetch(url, init)

  const isAuthRoute = path.startsWith('/auth/login') || path.startsWith('/auth/refresh')

  if (res.status === 401 && !isAuthRoute) {
    const ok = await doRefresh()
    if (ok) {
      const newToken = localStorage.getItem('sms_access_token')
      ;(init.headers as Record<string, string>).Authorization = `Bearer ${newToken}`
      res = await fetch(url, init)
    }
    if (res.status === 401) {
      localStorage.removeItem('sms_access_token')
      window.location.href = '/login'
      throw new ApiError(401, '登录已过期')
    }
  }

  if (res.status === 204) return undefined as T

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body.message || '请求失败')
  }

  return res.json()
}

export function buildQuery(params: Record<string, unknown>): string {
  const sp = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== '' && v !== undefined && v !== null) sp.set(k, String(v))
  }
  const qs = sp.toString()
  return qs ? `?${qs}` : ''
}

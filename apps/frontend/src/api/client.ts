import { API_BASE_PATH } from '@/config/app'
import { AUTHZ_REVISION_HEADER } from '@/features/session'
import { apiEvents } from './api-events'

export class ApiError extends Error {
  status: number
  retryAfter?: number

  constructor(status: number, message: string, retryAfter?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.retryAfter = retryAfter
  }
}

const NON_JSON_RESPONSE_MESSAGE =
  'Expected application/json from /api. Ensure /api is proxied to the Go backend (same-origin), not served as SPA HTML.'

function isNoContent(res: Response): boolean {
  return res.status === 204 || res.status === 205
}

function hasJsonContentType(res: Response): boolean {
  const ct = res.headers.get('Content-Type')
  return ct != null && ct.includes('application/json')
}

/**
 * Parse a successful response body.
 * - 204/205: no body expected → returns undefined.
 * - Otherwise: must be application/json (if not, the request likely hit
 *   the SPA fallback instead of the Go backend).
 */
async function parseSuccessBody<T>(res: Response): Promise<T> {
  if (isNoContent(res)) return undefined as T

  if (!hasJsonContentType(res)) {
    throw new ApiError(res.status, NON_JSON_RESPONSE_MESSAGE)
  }

  const text = await res.text()
  if (!text) return undefined as T

  try {
    return JSON.parse(text) as T
  } catch {
    throw new ApiError(res.status, 'Invalid JSON response from API')
  }
}

/** Best-effort parse of an error response body (may be JSON or empty). */
async function parseErrorBody(res: Response): Promise<{ message?: string; retryAfter?: number }> {
  if (isNoContent(res) || !hasJsonContentType(res)) return {}

  const text = await res.text()
  if (!text) return {}

  try {
    return JSON.parse(text)
  } catch {
    return {}
  }
}

function notifyAuthzRevision(res: Response): void {
  const revisionHeader = res.headers.get(AUTHZ_REVISION_HEADER)
  if (!revisionHeader) return
  const revision = Number(revisionHeader)
  if (Number.isFinite(revision)) {
    apiEvents.emit('authzRevision', revision)
  }
}

// Singleton refresh promise — concurrent 401s share one refresh call.
let refreshing: Promise<boolean> | null = null

function doRefresh(): Promise<boolean> {
  if (!refreshing) {
    refreshing = fetch(`${API_BASE_PATH}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
      .then((r) => r.ok)
      .catch(() => false)
      .finally(() => {
        refreshing = null
      })
  }
  return refreshing
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_PATH}${path}`
  const init: RequestInit = {
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  }

  let res = await fetch(url, init)
  notifyAuthzRevision(res)

  // 401 and not the refresh endpoint itself → attempt silent token refresh once
  if (res.status === 401 && path !== '/auth/refresh') {
    const ok = await doRefresh()
    if (ok) {
      res = await fetch(url, init)
      notifyAuthzRevision(res)
    }
  }

  if (!res.ok) {
    const body = await parseErrorBody(res)
    if (res.status === 401) {
      apiEvents.emit('unauthorized')
    }
    if (res.status === 403 && path !== '/session') {
      apiEvents.emit('forbidden', path)
    }
    throw new ApiError(
      res.status,
      body.message || res.statusText,
      typeof body.retryAfter === 'number' ? body.retryAfter : undefined,
    )
  }

  return parseSuccessBody<T>(res)
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

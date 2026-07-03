import { API_BASE_PATH } from '@/config/app'
import { AUTHZ_REVISION_HEADER } from '@/features/session/authz-sync'

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

let unauthorizedHandler: (() => void) | null = null
let forbiddenHandler: ((path: string) => void) | null = null
let authzRevisionHandler: ((revision: number) => void) | null = null

export function setUnauthorizedHandler(handler: (() => void) | null): void {
  unauthorizedHandler = handler
}

export function setForbiddenHandler(handler: ((path: string) => void) | null): void {
  forbiddenHandler = handler
}

export function setAuthzRevisionHandler(handler: ((revision: number) => void) | null): void {
  authzRevisionHandler = handler
}

const NON_JSON_RESPONSE_MESSAGE =
  'Expected application/json from /api. Ensure /api is proxied to the Go backend (same-origin), not served as SPA HTML.'

function isJsonContentType(contentType: string): boolean {
  return contentType.includes('application/json')
}

async function readJsonBody<T>(res: Response): Promise<T> {
  const contentType = res.headers.get('Content-Type') ?? ''
  if (!isJsonContentType(contentType)) {
    throw new ApiError(res.status, NON_JSON_RESPONSE_MESSAGE)
  }

  const text = await res.text()
  if (!text) {
    return undefined as T
  }

  try {
    return JSON.parse(text) as T
  } catch {
    throw new ApiError(res.status, 'Invalid JSON response from API')
  }
}

function notifyAuthzRevision(res: Response): void {
  const revisionHeader = res.headers.get(AUTHZ_REVISION_HEADER)
  if (!revisionHeader) return
  const revision = Number(revisionHeader)
  if (Number.isFinite(revision)) {
    authzRevisionHandler?.(revision)
  }
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_PATH}${path}`
  const res = await fetch(url, {
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  })

  notifyAuthzRevision(res)

  if (!res.ok) {
    let body: { message?: string; retryAfter?: number } = {}
    try {
      body = await readJsonBody<{ message?: string; retryAfter?: number }>(res)
    } catch (error) {
      if (error instanceof ApiError) {
        body = { message: error.message }
      }
    }
    if (res.status === 401) {
      unauthorizedHandler?.()
    }
    if (res.status === 403 && path !== '/session') {
      forbiddenHandler?.(path)
    }
    throw new ApiError(
      res.status,
      body.message || res.statusText,
      typeof body.retryAfter === 'number' ? body.retryAfter : undefined,
    )
  }

  return readJsonBody<T>(res)
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

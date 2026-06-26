const normalizedBase = import.meta.env.BASE_URL.replace(/\/$/, '')

export const API_BASE_PATH = `${normalizedBase}/api`

export const SERVICE_WORKER_URL = `${import.meta.env.BASE_URL}mockServiceWorker.js`
export const SERVICE_WORKER_SCOPE = import.meta.env.BASE_URL

export function resolveUseMocks(env: { DEV: boolean; VITE_ENABLE_MOCKS?: string }): boolean {
  return env.VITE_ENABLE_MOCKS === 'true' || (env.DEV && env.VITE_ENABLE_MOCKS !== 'false')
}

export const USE_MOCKS = resolveUseMocks({
  DEV: import.meta.env.DEV,
  VITE_ENABLE_MOCKS: import.meta.env.VITE_ENABLE_MOCKS,
})

export const API_PROXY_TARGET = import.meta.env.VITE_API_PROXY_TARGET
export const USE_API_PROXY = Boolean(API_PROXY_TARGET)

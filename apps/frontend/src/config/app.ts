const normalizedBase = import.meta.env.BASE_URL.replace(/\/$/, '')

export const API_BASE_PATH = `${normalizedBase}/api`

export const SERVICE_WORKER_URL = `${import.meta.env.BASE_URL}mockServiceWorker.js`
export const SERVICE_WORKER_SCOPE = import.meta.env.BASE_URL

export const USE_MOCKS = import.meta.env.DEV || import.meta.env.VITE_ENABLE_MOCKS === 'true'

export const API_PROXY_TARGET = import.meta.env.VITE_API_PROXY_TARGET

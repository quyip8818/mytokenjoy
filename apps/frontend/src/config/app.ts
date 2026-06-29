const normalizedBase = import.meta.env.BASE_URL.replace(/\/$/, '')

export const API_BASE_PATH = `${normalizedBase}/api`

export const API_PROXY_TARGET = import.meta.env.VITE_API_PROXY_TARGET

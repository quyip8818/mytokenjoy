const normalizedBase = import.meta.env.BASE_URL.replace(/\/$/, '')

export const API_BASE_PATH = `${normalizedBase}/api`

/**
 * Deploy mode: 'saas' enables registration/multi-tenant, 'local' is single-tenant (login only).
 * Set at build/start time via VITE_SUPPORT_SAAS env var.
 */
export const DEPLOY_MODE: 'saas' | 'local' =
  import.meta.env.VITE_SUPPORT_SAAS === 'true' ? 'saas' : 'local'

export const IS_SAAS = DEPLOY_MODE === 'saas'

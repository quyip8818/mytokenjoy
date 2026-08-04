export const LOGIN_PATH = '/login'

/** Paths that skip session fetch (no auth required). */
export const PUBLIC_PATHS = [LOGIN_PATH, '/invite/accept'] as const

export const SESSION_COOKIE = 'tokenjoy_session_member'
export const SESSION_COOKIE_MAX_AGE_DAYS = 7

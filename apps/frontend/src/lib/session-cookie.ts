import { SESSION_COOKIE, SESSION_COOKIE_MAX_AGE_DAYS } from '@/config/auth'

function cookiePath(): string {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  return base || '/'
}

export function setSessionMemberCookie(memberId: string): void {
  const maxAge = SESSION_COOKIE_MAX_AGE_DAYS * 24 * 60 * 60
  const encoded = encodeURIComponent(memberId)
  document.cookie = `${SESSION_COOKIE}=${encoded}; path=${cookiePath()}; max-age=${maxAge}; SameSite=Lax`
}

export function clearSessionMemberCookie(): void {
  document.cookie = `${SESSION_COOKIE}=; path=${cookiePath()}; max-age=0; SameSite=Lax`
}

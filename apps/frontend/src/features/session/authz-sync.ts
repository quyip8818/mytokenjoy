export const AUTHZ_BROADCAST_CHANNEL = 'tokenjoy-authz'
export const AUTHZ_REVISION_HEADER = 'x-authz-revision'
export const SESSION_FOCUS_REFRESH_MS = 60_000

export function broadcastAuthzChange(): void {
  if (typeof BroadcastChannel === 'undefined') return
  const channel = new BroadcastChannel(AUTHZ_BROADCAST_CHANNEL)
  channel.postMessage({ type: 'authz-changed' })
  channel.close()
}

export function parseAuthzRevision(headerValue: string | null): number | null {
  if (!headerValue) return null
  const parsed = Number(headerValue)
  return Number.isFinite(parsed) ? parsed : null
}

import type { ProxyOptions } from 'vite'

const BACKEND_API_PREFIX = '/api'
const DEFAULT_PROXY_TARGET = 'http://127.0.0.1:8080'

export function resolveApiPublicPrefix(base: string): string {
  const normalizedBase = base.replace(/\/$/, '')
  if (!normalizedBase) {
    return BACKEND_API_PREFIX
  }
  return `${normalizedBase}${BACKEND_API_PREFIX}`
}

function rewriteToBackendApi(publicPrefix: string): (path: string) => string {
  const escaped = publicPrefix.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const pattern = new RegExp(`^${escaped}`)
  return (path) => path.replace(pattern, BACKEND_API_PREFIX)
}

function proxyErrorHandler(
  _err: Error,
  _req: unknown,
  res: {
    writeHead?: (status: number, headers: Record<string, string>) => void
    end?: (body: string) => void
    headersSent?: boolean
  },
): void {
  if (res.writeHead && res.end && !res.headersSent) {
    res.writeHead(502, { 'Content-Type': 'application/json' })
    res.end(JSON.stringify({ message: 'API backend unavailable' }))
  }
}

export function createApiProxyConfig(base: string, target?: string): Record<string, ProxyOptions> {
  const publicPrefix = resolveApiPublicPrefix(base)
  const proxyTarget = target ?? DEFAULT_PROXY_TARGET

  return {
    [publicPrefix]: {
      target: proxyTarget,
      changeOrigin: true,
      rewrite: rewriteToBackendApi(publicPrefix),
      configure: (proxy) => {
        proxy.on('error', proxyErrorHandler)
      },
    },
  }
}

export function warnIfBackendUnreachable(target: string): void {
  const healthUrl = `${target.replace(/\/$/, '')}/healthz`
  void fetch(healthUrl)
    .then((res) => {
      if (!res.ok) {
        console.warn(
          `[tokenjoy] API backend not healthy at ${healthUrl} (HTTP ${res.status}). ` +
            'Start it with: pnpm start:backend',
        )
      }
    })
    .catch(() => {
      console.warn(
        `[tokenjoy] API backend unreachable at ${healthUrl}. ` +
          'Start it with: pnpm start:backend (or pnpm start from repo root)',
      )
    })
}

import { describe, expect, it } from 'vitest'
import { createApiProxyConfig, resolveApiPublicPrefix } from '../vite-api-proxy'

describe('resolveApiPublicPrefix', () => {
  it('uses /api at site root', () => {
    expect(resolveApiPublicPrefix('/')).toBe('/api')
  })

  it('prefixes API path with subpath base', () => {
    expect(resolveApiPublicPrefix('/mytokenjoy/')).toBe('/mytokenjoy/api')
  })
})

describe('createApiProxyConfig', () => {
  it('rewrites public prefix to backend /api', () => {
    const config = createApiProxyConfig('/mytokenjoy/', 'http://127.0.0.1:8080')
    const entry = config['/mytokenjoy/api']
    expect(entry?.rewrite?.('/mytokenjoy/api/session')).toBe('/api/session')
  })
})

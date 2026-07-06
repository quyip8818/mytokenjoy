import { describe, expect, it } from 'vitest'
import { createApiProxyConfig, resolveApiPublicPrefix } from '../vite-api-proxy'

describe('resolveApiPublicPrefix', () => {
  it('uses /api at site root', () => {
    expect(resolveApiPublicPrefix('/')).toBe('/api')
  })
})

describe('createApiProxyConfig', () => {
  it('rewrites public prefix to backend /api', () => {
    const config = createApiProxyConfig('/', 'http://127.0.0.1:8080')
    const entry = config['/api']
    expect(entry?.rewrite?.('/api/session')).toBe('/api/session')
  })
})

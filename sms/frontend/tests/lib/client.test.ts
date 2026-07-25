import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { buildQuery } from '@/api/client'

describe('buildQuery', () => {
  it('builds query string from params', () => {
    const qs = buildQuery({ page: 1, keyword: 'hello', status: 'active' })
    expect(qs).toContain('page=1')
    expect(qs).toContain('keyword=hello')
    expect(qs).toContain('status=active')
    expect(qs.startsWith('?')).toBe(true)
  })

  it('omits empty string, null, undefined values', () => {
    const qs = buildQuery({ page: 1, keyword: '', nothing: null, undef: undefined })
    expect(qs).toBe('?page=1')
  })

  it('returns empty string when all values are empty', () => {
    const qs = buildQuery({ a: '', b: null, c: undefined })
    expect(qs).toBe('')
  })

  it('stringifies numbers and booleans', () => {
    const qs = buildQuery({ count: 42, active: true })
    expect(qs).toContain('count=42')
    expect(qs).toContain('active=true')
  })

  it('handles zero as a valid value', () => {
    const qs = buildQuery({ page: 0 })
    // 0 is falsy but not empty/null/undefined — check implementation
    // buildQuery checks v !== '' && v !== undefined && v !== null
    expect(qs).toContain('page=0')
  })
})

describe('ApiError & request', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'localStorage',
      (() => {
        const store: Record<string, string> = { sms_access_token: 'test-token' }
        return {
          getItem: (k: string) => store[k] ?? null,
          setItem: (k: string, v: string) => {
            store[k] = v
          },
          removeItem: (k: string) => {
            delete store[k]
          },
          clear: () => Object.keys(store).forEach((k) => delete store[k]),
        }
      })(),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('ApiError has status and message', async () => {
    // dynamic import to get fresh module with stubbed globals
    const { ApiError } = await import('@/api/client')
    const err = new ApiError(404, 'not found')
    expect(err.status).toBe(404)
    expect(err.message).toBe('not found')
    expect(err.name).toBe('ApiError')
  })
})

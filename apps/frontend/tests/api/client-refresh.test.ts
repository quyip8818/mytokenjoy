import { afterEach, describe, expect, it, vi } from 'vitest'
import { request } from '@/api/client'

describe('api client 401 refresh retry', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('retries after successful refresh on 401', async () => {
    let callCount = 0
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
      const url = typeof input === 'string' ? input : (input as Request).url
      // refresh endpoint → success
      if (url.includes('/auth/refresh')) {
        return { ok: true, headers: new Headers() } as Response
      }
      // first call to /session → 401, second → 200
      callCount++
      if (callCount === 1) {
        return {
          ok: false,
          status: 401,
          statusText: 'Unauthorized',
          headers: new Headers({ 'Content-Type': 'application/json' }),
          text: async () => JSON.stringify({ message: 'token expired' }),
        } as Response
      }
      return {
        ok: true,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        text: async () => JSON.stringify({ data: 'ok' }),
      } as Response
    })

    const result = await request<{ data: string }>('/session')
    expect(result.data).toBe('ok')
    // 3 fetches: original 401, refresh, retry
    expect(globalThis.fetch).toHaveBeenCalledTimes(3)
  })

  it('emits unauthorized when refresh fails', async () => {
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
      const url = typeof input === 'string' ? input : (input as Request).url
      if (url.includes('/auth/refresh')) {
        return { ok: false, status: 401, headers: new Headers() } as Response
      }
      return {
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        headers: new Headers({ 'Content-Type': 'application/json' }),
        text: async () => JSON.stringify({ message: 'Unauthorized' }),
      } as Response
    })

    await expect(request('/session')).rejects.toMatchObject({ status: 401 })
    // 2 fetches: original 401, failed refresh (no retry)
    expect(globalThis.fetch).toHaveBeenCalledTimes(2)
  })

  it('does not refresh for /auth/refresh itself', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'Content-Type': 'application/json' }),
      text: async () => JSON.stringify({ message: 'Unauthorized' }),
    } as Response)

    await expect(request('/auth/refresh')).rejects.toMatchObject({ status: 401 })
    // Only 1 fetch — no refresh attempt
    expect(globalThis.fetch).toHaveBeenCalledTimes(1)
  })
})

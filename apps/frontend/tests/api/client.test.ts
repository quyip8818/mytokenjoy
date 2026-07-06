import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, request } from '@/api/client'

describe('api client request', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('parses successful JSON responses', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      text: async () => JSON.stringify({ status: 'ok' }),
    } as Response)

    const data = await request<{ status: string }>('/session')
    expect(data.status).toBe('ok')
  })

  it('parses JSON error bodies', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'Content-Type': 'application/json' }),
      text: async () => JSON.stringify({ message: 'Unauthorized' }),
    } as Response)

    await expect(request('/session')).rejects.toMatchObject({
      name: 'ApiError',
      status: 401,
      message: 'Unauthorized',
    })
  })

  it('rejects non-JSON responses', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers({ 'Content-Type': 'text/html' }),
      text: async () => '<!doctype html><html></html>',
    } as Response)

    await expect(request('/session')).rejects.toBeInstanceOf(ApiError)
    await expect(request('/session')).rejects.toMatchObject({
      message: expect.stringContaining('application/json'),
    })
  })

  it('sends Accept application/json', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      text: async () => 'null',
    } as Response)

    await request('/session')

    const [, init] = fetchMock.mock.calls[0]!
    expect(init?.headers).toMatchObject({
      Accept: 'application/json',
      'Content-Type': 'application/json',
    })
  })
})

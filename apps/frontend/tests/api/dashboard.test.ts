import { afterEach, describe, expect, it, vi } from 'vitest'
import { dashboardApi } from '@/api/dashboard'

describe('dashboardApi.getUsageSeries', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('requests usage series with query params', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: async () => ({
        granularity: 'hour',
        source: 'buckets',
        timezone: 'Asia/Shanghai',
        approximate: false,
        mappingAsOf: 'ingest_time',
        points: [],
      }),
    } as Response)

    await dashboardApi.getUsageSeries({
      granularity: 'hour',
      start: '2026-06-01T00:00:00.000Z',
      end: '2026-06-02T00:00:00.000Z',
      groupBy: 'none',
    })

    expect(fetchMock).toHaveBeenCalledOnce()
    const [url] = fetchMock.mock.calls[0]
    expect(String(url)).toContain('/dashboard/usage/series?')
    expect(String(url)).toContain('granularity=hour')
    expect(String(url)).toContain('groupBy=none')
  })
})

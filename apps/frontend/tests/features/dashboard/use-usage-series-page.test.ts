import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useUsageSeriesPage } from '@/features/dashboard/hooks/use-usage-series-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useUsageSeriesPage', () => {
  it('loads usage series with approximate metadata', async () => {
    const getUsageSeries = vi.fn().mockResolvedValue({
      granularity: 'minute',
      source: 'ledger',
      timezone: 'Asia/Shanghai',
      approximate: true,
      mappingAsOf: 'query_time',
      points: [
        {
          bucket: '2026-06-10T10:05:00+08:00',
          costCny: 3,
          callCount: 1,
          inputTokens: 0,
          outputTokens: 0,
        },
      ],
    })
    const apis = createMockApis({
      dashboardApi: { getUsageSeries },
    })
    const { result } = renderHookWithProviders(() => useUsageSeriesPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(getUsageSeries).toHaveBeenCalledWith(
      expect.objectContaining({
        granularity: 'hour',
        groupBy: 'none',
      }),
    )
    expect(result.current.series?.approximate).toBe(true)
    expect(result.current.chartData).toHaveLength(1)
  })

  it('switches granularity and refetches', async () => {
    const getUsageSeries = vi.fn().mockResolvedValue({
      granularity: 'hour',
      source: 'buckets',
      timezone: 'Asia/Shanghai',
      approximate: false,
      mappingAsOf: 'ingest_time',
      points: [],
    })
    const apis = createMockApis({
      dashboardApi: { getUsageSeries },
    })
    const { result } = renderHookWithProviders(() => useUsageSeriesPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.handleGranularityChange('hour')
    })

    await waitForLoaded(result, 'loading')

    expect(getUsageSeries).toHaveBeenLastCalledWith(
      expect.objectContaining({ granularity: 'hour' }),
    )
  })
})

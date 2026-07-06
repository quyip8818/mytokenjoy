import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { UsageGranularity, UsageSeriesQuery } from '@/api/types'
import { ApiError } from '@/api/client'
import { USAGE_GRANULARITY } from '@/lib/dashboard-constants'
import { buildUsageSeriesChartData, buildUsageSeriesWindow } from '@/lib/dashboard'
import { queryKeys, useInjectedQuery } from '@/features/query'

function buildQuery(granularity: UsageGranularity): UsageSeriesQuery {
  const window = buildUsageSeriesWindow(granularity)
  return {
    granularity,
    start: window.start,
    end: window.end,
    groupBy: 'none',
  }
}

export function useUsageSeriesPage(injectedApis?: AppApis) {
  const [granularity, setGranularity] = useState<UsageGranularity>(USAGE_GRANULARITY.HOUR)
  const seriesQuery = useMemo(() => buildQuery(granularity), [granularity])

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.dashboard.usageSeries(seriesQuery),
    queryFn: (apis) => apis.dashboardApi.getUsageSeries(seriesQuery),
  })

  const chartGranularity = data?.granularity ?? granularity
  const chartData = useMemo(
    () => buildUsageSeriesChartData(data?.points ?? [], chartGranularity),
    [data?.points, chartGranularity],
  )

  const isServiceUnavailable = error instanceof ApiError && error.status === 503
  const serviceUnavailableMessage = isServiceUnavailable ? 'NewAPI 日志暂不可用，请稍后重试' : null

  const handleGranularityChange = useCallback((value: string | null) => {
    if (!value) return
    setGranularity(value as UsageGranularity)
  }, [])

  return {
    granularity,
    seriesQuery,
    series: data ?? null,
    chartData,
    loading,
    error,
    isServiceUnavailable,
    serviceUnavailableMessage,
    retryAfter: error instanceof ApiError ? error.retryAfter : undefined,
    refresh,
    handleGranularityChange,
  }
}

import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { CostPeriod, CostQueryParams } from '@/api/types'
import { COST_PERIOD } from '../lib/constants'
import { formatLocalDate, getMonthStartLocal, getTodayLocal, getWeekStartLocal } from '@/lib/date'
import { queryKeys, useInjectedQuery } from '@/features/query'

interface UseUsageDashboardPageOptions {
  deptId: string | null
  injectedApis?: AppApis
}

function buildCostQuery(period: CostPeriod, startDate: string, endDate: string): CostQueryParams {
  if (period === COST_PERIOD.LAST_30_DAYS) {
    const to = new Date()
    const from = new Date()
    from.setDate(from.getDate() - 29)
    return { period: 'custom', startDate: formatLocalDate(from), endDate: formatLocalDate(to) }
  }
  if (period === COST_PERIOD.CURRENT_WEEK) {
    return { period: 'custom', startDate: getWeekStartLocal(), endDate: getTodayLocal() }
  }
  if (period === COST_PERIOD.CUSTOM) {
    return { period, startDate, endDate }
  }
  return { period }
}

export function useUsageDashboardPage({ deptId, injectedApis }: UseUsageDashboardPageOptions) {
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [startDate, setStartDate] = useState(getMonthStartLocal)
  const [endDate, setEndDate] = useState(getTodayLocal)

  const costQuery = useMemo(
    () => buildCostQuery(period, startDate, endDate),
    [period, startDate, endDate],
  )

  const {
    data: teamUsage = [],
    loading: teamLoading,
    error: teamError,
    refresh: refreshTeam,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'team', costQuery, deptId],
    queryFn: (a) =>
      a.dashboardApi.getTeamUsage({ ...costQuery, departmentId: deptId ?? undefined }),
  })

  const {
    data: modelUsage = [],
    loading: modelLoading,
    error: modelError,
    refresh: refreshModel,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'model', costQuery, deptId],
    queryFn: (a) =>
      a.dashboardApi.getModelUsage({ ...costQuery, departmentId: deptId ?? undefined }),
  })

  const {
    data: topConsumers = [],
    loading: topLoading,
    error: topError,
    refresh: refreshTop,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'top', costQuery, deptId],
    queryFn: (a) =>
      a.dashboardApi.getTopConsumers({
        ...costQuery,
        limit: 10,
        departmentId: deptId ?? undefined,
      }),
  })

  const loading = teamLoading || modelLoading || topLoading
  const error = teamError ?? modelError ?? topError
  const refresh = async () => {
    await Promise.all([refreshTeam(), refreshModel(), refreshTop()])
  }

  const handlePeriodChange = useCallback((value: string | null) => {
    if (!value) return
    setPeriod(value as CostPeriod)
  }, [])

  const customDateInvalid =
    period === COST_PERIOD.CUSTOM && Boolean(startDate && endDate && startDate > endDate)

  return {
    period,
    startDate,
    endDate,
    customDateInvalid,
    deptId,
    teamUsage,
    modelUsage,
    topConsumers,
    loading,
    error,
    refresh,
    handlePeriodChange,
    setStartDate,
    setEndDate,
  }
}

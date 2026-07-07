import { useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { CostQueryParams } from '@/api/types'
import { COST_PERIOD } from '@/features/dashboard/lib/constants'
import { queryKeys, useInjectedQuery } from '@/features/query'

export function useUsageDashboardPage(injectedApis?: AppApis) {
  const costQuery = useMemo<CostQueryParams>(() => ({ period: COST_PERIOD.CURRENT_MONTH }), [])

  const {
    data: teamUsage = [],
    loading: teamLoading,
    error: teamError,
    refresh: refreshTeam,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'team', costQuery],
    queryFn: (a) => a.dashboardApi.getTeamUsage(costQuery),
  })
  const {
    data: modelUsage = [],
    loading: modelLoading,
    error: modelError,
    refresh: refreshModel,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'model', costQuery],
    queryFn: (a) => a.dashboardApi.getModelUsage(costQuery),
  })

  const loading = teamLoading || modelLoading
  const error = teamError ?? modelError
  const refresh = async () => {
    await Promise.all([refreshTeam(), refreshModel()])
  }

  return {
    teamUsage,
    modelUsage,
    loading,
    error,
    refresh,
  }
}

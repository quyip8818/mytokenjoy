import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'

export function useUsageDashboardPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const {
    data: teamUsage = [],
    loading: teamLoading,
    error: teamError,
    refresh: refreshTeam,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey: [...queryKeys.dashboard.usage(), 'team'],
    queryFn: (a) => a.dashboardApi.getTeamUsage(),
  })
  const {
    data: modelUsage = [],
    loading: modelLoading,
    error: modelError,
    refresh: refreshModel,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey: [...queryKeys.dashboard.usage(), 'model'],
    queryFn: (a) => a.dashboardApi.getModelUsage(),
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

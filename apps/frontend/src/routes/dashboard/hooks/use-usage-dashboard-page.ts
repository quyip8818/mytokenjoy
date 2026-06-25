import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useUsageDashboardPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { data, loading, error, refresh } = useAsyncResource(async () => {
    const [teamUsage, modelUsage] = await Promise.all([
      apis.dashboardApi.getTeamUsage(),
      apis.dashboardApi.getModelUsage(),
    ])
    return { teamUsage, modelUsage }
  }, [apis])

  return {
    teamUsage: data?.teamUsage ?? [],
    modelUsage: data?.modelUsage ?? [],
    loading,
    error,
    refresh,
  }
}

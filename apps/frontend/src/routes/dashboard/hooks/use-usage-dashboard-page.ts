import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useUsageDashboardPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
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

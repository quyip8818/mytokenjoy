import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'

export function useAuditModelOptions(injectedApis?: AppApis) {
  const { data: models = [], loading } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.audit.models(),
    queryFn: (apis) => apis.modelsApi.list(),
  })

  return { models, loading }
}

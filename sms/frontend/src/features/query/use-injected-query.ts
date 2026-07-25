import { useQuery, type QueryKey } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'

export function useInjectedQuery<T>({
  injectedApis,
  queryKey,
  queryFn,
  enabled,
}: {
  injectedApis?: AppApis
  queryKey: QueryKey
  queryFn: (apis: AppApis) => Promise<T>
  enabled?: boolean
}) {
  const apis = useInjectedApis(injectedApis)
  const query = useQuery({ queryKey, queryFn: () => queryFn(apis), enabled })
  return {
    data: query.data,
    loading: query.isLoading,
    error: query.error,
    refresh: async () => {
      await query.refetch()
    },
  }
}

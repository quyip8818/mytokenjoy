import { useQuery, type QueryKey, type UseQueryOptions } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'

type QueryFn<T> = (apis: AppApis) => Promise<T>

export interface UseInjectedQueryOptions<T> extends Omit<
  UseQueryOptions<T, Error, T, QueryKey>,
  'queryKey' | 'queryFn'
> {
  injectedApis?: AppApis
  queryKey: QueryKey
  queryFn: QueryFn<T>
}

export interface InjectedQueryViewModel<T> {
  data: T | undefined
  loading: boolean
  error: Error | null
  refresh: () => Promise<void>
}

export function useInjectedQuery<T>({
  injectedApis,
  queryKey,
  queryFn,
  ...options
}: UseInjectedQueryOptions<T>): InjectedQueryViewModel<T> {
  const apis = useInjectedApis(injectedApis)
  const query = useQuery({
    queryKey,
    queryFn: () => queryFn(apis),
    ...options,
  })

  const refresh = async () => {
    await query.refetch()
  }

  return {
    data: query.data,
    loading: query.isLoading,
    error: query.error,
    refresh,
  }
}

import { useCallback, useState } from 'react'
import type { QueryKey } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import { useInjectedQuery } from './use-injected-query'

export function useFilteredQuery<T, F extends { page?: number }>({
  injectedApis,
  initialFilter,
  queryKeyFactory,
  fetcher,
}: {
  injectedApis?: AppApis
  initialFilter: F
  queryKeyFactory: (filter: F) => QueryKey
  fetcher: (apis: AppApis, filter: F) => Promise<T>
}) {
  const [filter, setFilterState] = useState(initialFilter)
  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeyFactory(filter),
    queryFn: (apis) => fetcher(apis, filter),
  })
  const setFilter = useCallback((next: F | ((prev: F) => F)) => {
    setFilterState(next)
  }, [])
  const setPage = useCallback((page: number) => {
    setFilterState((prev) => ({ ...prev, page }))
  }, [])
  const search = useCallback((patch: Partial<F>) => {
    setFilterState((prev) => ({ ...prev, ...patch, page: 1 }))
  }, [])
  return { data, loading, error, refresh, filter, setFilter, setPage, search }
}

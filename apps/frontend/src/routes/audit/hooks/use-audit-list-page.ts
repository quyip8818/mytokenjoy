import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { useFilteredQuery } from '@/hooks/use-filtered-query'

export interface UseAuditListPageConfig<TFilter, TItem, TQuery> {
  initialFilter: TFilter
  toQueryParams: (filter: TFilter) => TQuery
  fetchItems: (apis: AppApis, query: TQuery) => Promise<TItem[]>
  injectedApis?: AppApis
  queryKeyFactory: (filter: TFilter) => readonly unknown[]
}

export function useAuditListPage<TFilter, TItem, TQuery>({
  initialFilter,
  toQueryParams,
  fetchItems,
  injectedApis,
  queryKeyFactory,
}: UseAuditListPageConfig<TFilter, TItem, TQuery>) {
  const apis = useInjectedApis(injectedApis)

  const { data, loading, error, refresh, filter, setFilter } = useFilteredQuery({
    injectedApis: apis,
    initialFilter,
    queryKeyFactory: (currentFilter) => queryKeyFactory(currentFilter),
    fetcher: (a, currentFilter) => fetchItems(a, toQueryParams(currentFilter)),
  })

  const patchFilter = useCallback(
    (patch: Partial<TFilter>) => {
      setFilter((prev) => ({ ...prev, ...patch }))
    },
    [setFilter],
  )

  return {
    items: data ?? [],
    filter,
    setFilter,
    patchFilter,
    loading,
    error,
    refresh,
  }
}

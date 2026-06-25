import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useFilteredResource } from '@/hooks/use-filtered-resource'

export interface UseAuditListPageConfig<TFilter, TItem, TQuery> {
  initialFilter: TFilter
  toQueryParams: (filter: TFilter) => TQuery
  fetchItems: (apis: AppApis, query: TQuery) => Promise<TItem[]>
  injectedApis?: AppApis
}

export function useAuditListPage<TFilter, TItem, TQuery>({
  initialFilter,
  toQueryParams,
  fetchItems,
  injectedApis,
}: UseAuditListPageConfig<TFilter, TItem, TQuery>) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis

  const { data, loading, error, refresh, filter, setFilter } = useFilteredResource(
    (currentFilter) => fetchItems(apis, toQueryParams(currentFilter)),
    initialFilter,
  )

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

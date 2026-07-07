import { useCallback, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { Paginated } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { AUDIT_PAGE_SIZE } from '../lib/constants'
import { useInjectedQuery } from '@/features/query/use-injected-query'

export interface UseAuditListPageConfig<TFilter, TItem, TQuery> {
  initialFilter: TFilter
  toQueryParams: (filter: TFilter) => TQuery
  fetchPage: (
    apis: AppApis,
    query: TQuery & { page: number; pageSize: number },
  ) => Promise<Paginated<TItem>>
  injectedApis?: AppApis
  queryKeyFactory: (params: { filter: TFilter; page: number }) => readonly unknown[]
  pageSize?: number
}

export function useAuditListPage<TFilter, TItem, TQuery>({
  initialFilter,
  toQueryParams,
  fetchPage,
  injectedApis,
  queryKeyFactory,
  pageSize = AUDIT_PAGE_SIZE,
}: UseAuditListPageConfig<TFilter, TItem, TQuery>) {
  const apis = useInjectedApis(injectedApis)
  const [filter, setFilterState] = useState(initialFilter)
  const [page, setPage] = useState(1)

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeyFactory({ filter, page }),
    queryFn: (a) =>
      fetchPage(a, {
        ...toQueryParams(filter),
        page,
        pageSize,
      }),
  })

  const setFilter = useCallback((next: TFilter | ((prev: TFilter) => TFilter)) => {
    setFilterState(next)
    setPage(1)
  }, [])

  const patchFilter = useCallback(
    (patch: Partial<TFilter>) => {
      setFilter((prev) => ({ ...prev, ...patch }))
    },
    [setFilter],
  )

  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return {
    items: data?.items ?? [],
    total,
    page,
    pageSize,
    totalPages,
    setPage,
    filter,
    setFilter,
    patchFilter,
    loading,
    error,
    refresh,
  }
}

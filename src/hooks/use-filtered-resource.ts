import { useCallback, useState } from 'react'
import { useAsyncResource } from './use-async-resource'

export function useFilteredResource<T, F>(
  fetcher: (filter: F, signal: AbortSignal) => Promise<T>,
  initialFilter: F,
) {
  const [filter, setFilterState] = useState(initialFilter)
  const { data, loading, error, refresh, setData, setLoading } = useAsyncResource(
    (signal) => fetcher(filter, signal),
    [filter],
  )

  const setFilter = useCallback(
    (next: F | ((prev: F) => F)) => {
      setFilterState(next)
      setLoading(true)
    },
    [setLoading],
  )

  return { data, loading, error, refresh, setData, filter, setFilter }
}

import { useEffect, useState } from 'react'

type AsyncFetchState<T> = {
  key: string
  loading: boolean
  data: T
}

export function useAsyncFetch<T>(
  key: string,
  fetch: () => Promise<T>,
  enabled = true,
  initialData: T,
) {
  const [state, setState] = useState<AsyncFetchState<T>>({
    key: '',
    loading: false,
    data: initialData,
  })

  useEffect(() => {
    if (!enabled || !key) return

    let cancelled = false
    fetch()
      .then((data) => {
        if (!cancelled) {
          setState({ key, loading: false, data })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setState({ key, loading: false, data: initialData })
        }
      })

    return () => {
      cancelled = true
    }
  }, [enabled, fetch, initialData, key])

  const loading = enabled && key !== '' && (state.key !== key || state.loading)
  const data = state.key === key ? state.data : initialData

  const refresh = async () => {
    const next = await fetch()
    setState({ key, loading: false, data: next })
    return next
  }

  const replace = (next: T) => {
    setState({ key, loading: false, data: next })
  }

  return { loading, data, refresh, replace }
}

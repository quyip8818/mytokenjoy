import { useCallback, useEffect, useState, type Dispatch, type SetStateAction } from 'react'

export type AsyncFetcher<T> = (signal: AbortSignal) => Promise<T>

export interface UseAsyncResourceResult<T> {
  data: T | undefined
  loading: boolean
  error: Error | null
  refresh: () => Promise<void>
  setData: Dispatch<SetStateAction<T | undefined>>
  setLoading: Dispatch<SetStateAction<boolean>>
}

export function useAsyncResource<T>(
  fetcher: AsyncFetcher<T>,
  deps: unknown[] = [],
): UseAsyncResourceResult<T> {
  const [data, setData] = useState<T | undefined>(undefined)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await fetcher(new AbortController().signal)
      setData(result)
    } catch (e) {
      setError(e instanceof Error ? e : new Error(String(e)))
    } finally {
      setLoading(false)
    }
  }, [fetcher])

  useEffect(() => {
    let cancelled = false
    const controller = new AbortController()

    queueMicrotask(() => {
      if (!cancelled) {
        setLoading(true)
        setError(null)
      }
    })

    void fetcher(controller.signal)
      .then((result) => {
        if (!cancelled) {
          setData(result)
          setLoading(false)
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setError(e instanceof Error ? e : new Error(String(e)))
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
      controller.abort()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fetcher, ...deps])

  return { data, loading, error, refresh, setData, setLoading }
}

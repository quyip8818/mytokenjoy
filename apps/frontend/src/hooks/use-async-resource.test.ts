import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useAsyncResource } from './use-async-resource'

describe('useAsyncResource', () => {
  it('loads data on mount and exposes refresh', async () => {
    const fetcher = vi.fn(async () => 'loaded')

    const { result } = renderHook(() => useAsyncResource(fetcher, []))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(fetcher).toHaveBeenCalledWith(expect.any(AbortSignal))
    expect(result.current.data).toBe('loaded')
    expect(result.current.error).toBeNull()

    fetcher.mockResolvedValueOnce('refreshed')
    await act(async () => {
      await result.current.refresh()
    })

    expect(result.current.data).toBe('refreshed')
  })

  it('captures fetch errors', async () => {
    const fetcher = vi.fn(async () => {
      throw new Error('fetch failed')
    })

    const { result } = renderHook(() => useAsyncResource(fetcher, []))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toBeUndefined()
    expect(result.current.error?.message).toBe('fetch failed')
  })

  it('re-fetches when deps change', async () => {
    const fetcher = vi.fn(async (signal: AbortSignal) => {
      return `value:${signal.aborted}`
    })

    const { result, rerender } = renderHook(
      ({ dep }) => useAsyncResource(() => fetcher(new AbortController().signal), [dep]),
      { initialProps: { dep: 1 } },
    )

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    rerender({ dep: 2 })

    await waitFor(() => {
      expect(fetcher).toHaveBeenCalledTimes(2)
    })
  })
})

import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useFilteredResource } from '@/hooks/use-filtered-resource'

describe('useFilteredResource', () => {
  it('sets loading when filter changes and resolves with new data', async () => {
    const fetcher = vi.fn(async (filter: string) => `result:${filter}`)

    const { result } = renderHook(() => useFilteredResource<string, string>(fetcher, 'a'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    expect(result.current.data).toBe('result:a')

    act(() => {
      result.current.setFilter('b')
    })

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    expect(fetcher).toHaveBeenLastCalledWith('b', expect.any(AbortSignal))
    expect(result.current.data).toBe('result:b')
  })
})

import { describe, it, expect, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useFilteredQuery } from '@/features/query/use-filtered-query'
import { createTestWrapper, createMockApis } from '@tests/utils'

interface Filter {
  page?: number
  keyword?: string
}

const fetcher = vi.fn()

function setup(initialFilter: Filter = { page: 1, keyword: '' }) {
  fetcher.mockResolvedValue({ items: ['a', 'b'], total: 2, page: 1, pageSize: 10 })
  const apis = createMockApis()
  return renderHook(
    () =>
      useFilteredQuery<{ items: string[]; total: number }, Filter>({
        injectedApis: apis,
        initialFilter,
        queryKeyFactory: (f) => ['test', f],
        fetcher: (_apis, f) => fetcher(f),
      }),
    { wrapper: createTestWrapper({ apis }) },
  )
}

describe('useFilteredQuery', () => {
  it('fetches with initial filter', async () => {
    const { result } = setup()
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(fetcher).toHaveBeenCalledWith({ page: 1, keyword: '' })
    expect(result.current.data).toEqual({ items: ['a', 'b'], total: 2, page: 1, pageSize: 10 })
  })

  it('setPage updates page in filter', async () => {
    const { result } = setup()
    await waitFor(() => expect(result.current.loading).toBe(false))

    fetcher.mockResolvedValue({ items: ['c'], total: 2, page: 2, pageSize: 10 })
    act(() => result.current.setPage(2))

    expect(result.current.filter.page).toBe(2)
    await waitFor(() => expect(fetcher).toHaveBeenCalledWith({ page: 2, keyword: '' }))
  })

  it('search resets page to 1 and merges patch', async () => {
    const { result } = setup({ page: 3, keyword: '' })
    await waitFor(() => expect(result.current.loading).toBe(false))

    fetcher.mockResolvedValue({ items: ['x'], total: 1, page: 1, pageSize: 10 })
    act(() => result.current.search({ keyword: 'hello' }))

    expect(result.current.filter.page).toBe(1)
    expect(result.current.filter.keyword).toBe('hello')
    await waitFor(() => expect(fetcher).toHaveBeenCalledWith({ page: 1, keyword: 'hello' }))
  })

  it('setFilter allows full replacement', async () => {
    const { result } = setup()
    await waitFor(() => expect(result.current.loading).toBe(false))

    fetcher.mockResolvedValue({ items: [], total: 0, page: 5, pageSize: 10 })
    act(() => result.current.setFilter({ page: 5, keyword: 'new' }))

    expect(result.current.filter).toEqual({ page: 5, keyword: 'new' })
  })
})

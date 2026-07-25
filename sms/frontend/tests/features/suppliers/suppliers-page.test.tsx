import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor, fireEvent } from '@testing-library/react'
import { SuppliersPage } from '@/features/suppliers'
import { renderWithProviders, createMockApis } from '@tests/utils'
import { useSession } from '@/features/session'

const mockSuppliersResponse = {
  items: [
    {
      id: 's1',
      name: 'OpenAI',
      code: 'OPENAI',
      category: '国外厂商',
      website: 'https://openai.com',
      status: 'active',
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01',
    },
    {
      id: 's2',
      name: '智谱AI',
      code: 'ZHIPU',
      category: '国内厂商',
      status: 'potential',
      createdAt: '2024-01-02',
      updatedAt: '2024-01-02',
    },
  ],
  total: 2,
  page: 1,
  pageSize: 10,
}

describe('SuppliersPage', () => {
  beforeEach(() => {
    useSession.setState({
      user: { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' },
    })
  })

  it('renders supplier list from API', async () => {
    const apis = createMockApis({
      suppliersApi: { list: vi.fn().mockResolvedValue(mockSuppliersResponse) },
    })

    renderWithProviders(<SuppliersPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('OpenAI')).toBeInTheDocument()
    })
    expect(screen.getByText('智谱AI')).toBeInTheDocument()
    expect(screen.getByText('OPENAI')).toBeInTheDocument()
  })

  it('shows 新建供应商 button for admin', async () => {
    const apis = createMockApis({
      suppliersApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
    })

    renderWithProviders(<SuppliersPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('新建供应商')).toBeInTheDocument()
    })
  })

  it('hides 新建供应商 button for viewer', async () => {
    useSession.setState({
      user: { id: 'u2', username: 'viewer', realName: '查看者', role: 'viewer' },
    })
    const apis = createMockApis({
      suppliersApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
    })

    renderWithProviders(<SuppliersPage />, { apis })

    await waitFor(() => {
      expect(screen.queryByText('新建供应商')).not.toBeInTheDocument()
    })
  })

  it('filters by keyword on input change', async () => {
    const listFn = vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
    const apis = createMockApis({
      suppliersApi: { list: listFn },
    })

    renderWithProviders(<SuppliersPage />, { apis })

    await waitFor(() => expect(listFn).toHaveBeenCalled())

    const input = screen.getByPlaceholderText('名称 / 编码')
    fireEvent.change(input, { target: { value: 'Open' } })

    await waitFor(() => {
      expect(listFn).toHaveBeenCalledWith(expect.objectContaining({ keyword: 'Open', page: 1 }))
    })
  })

  it('shows empty state when no suppliers', async () => {
    const apis = createMockApis({
      suppliersApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
    })

    renderWithProviders(<SuppliersPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('暂无数据')).toBeInTheDocument()
    })
  })
})

describe('useSupplierOptions', () => {
  it('returns options from API', async () => {
    const { renderHookWithProviders } = await import('@tests/utils')
    const { useSupplierOptions } = await import('@/features/suppliers')
    const apis = createMockApis({
      suppliersApi: {
        options: vi.fn().mockResolvedValue([
          { id: 's1', name: 'OpenAI' },
          { id: 's2', name: '智谱AI' },
        ]),
      },
    })

    const { result } = renderHookWithProviders(() => useSupplierOptions(), { apis })

    await waitFor(() => {
      expect(result.current.length).toBe(2)
    })
    expect(result.current[0].name).toBe('OpenAI')
  })

  it('returns empty array when no data', async () => {
    const { renderHookWithProviders } = await import('@tests/utils')
    const { useSupplierOptions } = await import('@/features/suppliers')
    const apis = createMockApis({
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    const { result } = renderHookWithProviders(() => useSupplierOptions(), { apis })

    await waitFor(() => {
      expect(result.current).toEqual([])
    })
  })
})

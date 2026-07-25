import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor, fireEvent } from '@testing-library/react'
import { ContractsPage } from '@/features/contracts'
import { renderWithProviders, createMockApis } from '@tests/utils'
import { useSession } from '@/features/session'

const mockContractsResponse = {
  items: [
    {
      id: 'c1',
      supplierId: 's1',
      contractNo: 'HT-001',
      title: 'GPT-4 服务合同',
      amount: 50000,
      startDate: '2024-01-01',
      endDate: '2025-12-31',
      status: 'active',
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01',
      supplierName: 'OpenAI',
    },
    {
      id: 'c2',
      supplierId: 's2',
      contractNo: 'HT-002',
      title: 'Claude 合作协议',
      status: 'draft',
      createdAt: '2024-02-01',
      updatedAt: '2024-02-01',
      supplierName: 'Anthropic',
    },
  ],
  total: 2,
  page: 1,
  pageSize: 10,
}

describe('ContractsPage', () => {
  beforeEach(() => {
    useSession.setState({
      user: { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' },
    })
  })

  it('renders contract list from API', async () => {
    const apis = createMockApis({
      contractsApi: { list: vi.fn().mockResolvedValue(mockContractsResponse) },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<ContractsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('GPT-4 服务合同')).toBeInTheDocument()
    })
    expect(screen.getByText('Claude 合作协议')).toBeInTheDocument()
    expect(screen.getByText('HT-001')).toBeInTheDocument()
    expect(screen.getByText('OpenAI')).toBeInTheDocument()
  })

  it('shows 新建合同 button for admin', async () => {
    const apis = createMockApis({
      contractsApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<ContractsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('新建合同')).toBeInTheDocument()
    })
  })

  it('hides 新建合同 button for viewer', async () => {
    useSession.setState({
      user: { id: 'u2', username: 'viewer', realName: '查看者', role: 'viewer' },
    })
    const apis = createMockApis({
      contractsApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<ContractsPage />, { apis })

    await waitFor(() => {
      expect(screen.queryByText('新建合同')).not.toBeInTheDocument()
    })
  })

  it('filters by keyword on input change', async () => {
    const listFn = vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
    const apis = createMockApis({
      contractsApi: { list: listFn },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<ContractsPage />, { apis })

    await waitFor(() => expect(listFn).toHaveBeenCalled())

    const input = screen.getByPlaceholderText('合同编号 / 标题')
    fireEvent.change(input, { target: { value: 'HT-X' } })

    await waitFor(() => {
      expect(listFn).toHaveBeenCalledWith(expect.objectContaining({ keyword: 'HT-X', page: 1 }))
    })
  })

  it('shows empty state when no contracts', async () => {
    const apis = createMockApis({
      contractsApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<ContractsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('暂无数据')).toBeInTheDocument()
    })
  })

  it('displays formatted amount', async () => {
    const apis = createMockApis({
      contractsApi: { list: vi.fn().mockResolvedValue(mockContractsResponse) },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<ContractsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('50,000.00')).toBeInTheDocument()
    })
  })
})

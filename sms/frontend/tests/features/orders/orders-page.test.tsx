import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor, fireEvent } from '@testing-library/react'
import { OrdersPage } from '@/features/orders'
import { renderWithProviders, createMockApis } from '@tests/utils'
import { useSession } from '@/features/session'

const mockOrdersResponse = {
  items: [
    {
      id: 'o1',
      orderNo: 'PO-001',
      supplierId: 's1',
      totalAmount: 10000,
      orderDate: '2024-06-01',
      status: 'pending',
      createdAt: '2024-06-01',
      updatedAt: '2024-06-01',
      supplierName: 'Acme',
      contractNo: 'CT-001',
      creatorName: '管理员',
    },
    {
      id: 'o2',
      orderNo: 'PO-002',
      supplierId: 's1',
      totalAmount: 5000,
      status: 'approved',
      createdAt: '2024-06-02',
      updatedAt: '2024-06-02',
      supplierName: 'Acme',
      creatorName: '管理员',
    },
  ],
  total: 2,
  page: 1,
  pageSize: 10,
}

describe('OrdersPage', () => {
  beforeEach(() => {
    useSession.setState({
      user: { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' },
    })
  })

  it('renders order list from API', async () => {
    const apis = createMockApis({
      ordersApi: {
        list: vi.fn().mockResolvedValue(mockOrdersResponse),
      },
      suppliersApi: {
        options: vi.fn().mockResolvedValue([]),
      },
    })

    renderWithProviders(<OrdersPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('PO-001')).toBeInTheDocument()
    })
    expect(screen.getByText('PO-002')).toBeInTheDocument()
    expect(screen.getAllByText('Acme').length).toBeGreaterThan(0)
  })

  it('shows 新建订单 button for admin', async () => {
    const apis = createMockApis({
      ordersApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<OrdersPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('新建订单')).toBeInTheDocument()
    })
  })

  it('hides 新建订单 button for viewer', async () => {
    useSession.setState({
      user: { id: 'u2', username: 'viewer', realName: '查看者', role: 'viewer' },
    })
    const apis = createMockApis({
      ordersApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<OrdersPage />, { apis })

    await waitFor(() => {
      expect(screen.queryByText('新建订单')).not.toBeInTheDocument()
    })
  })

  it('calls list API with filter params on search', async () => {
    const listFn = vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
    const apis = createMockApis({
      ordersApi: { list: listFn },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<OrdersPage />, { apis })

    await waitFor(() => expect(listFn).toHaveBeenCalled())

    const input = screen.getByPlaceholderText('订单编号')
    fireEvent.change(input, { target: { value: 'PO-X' } })

    await waitFor(() => {
      expect(listFn).toHaveBeenCalledWith(expect.objectContaining({ keyword: 'PO-X', page: 1 }))
    })
  })

  it('shows empty state when no orders', async () => {
    const apis = createMockApis({
      ordersApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<OrdersPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('暂无数据')).toBeInTheDocument()
    })
  })
})

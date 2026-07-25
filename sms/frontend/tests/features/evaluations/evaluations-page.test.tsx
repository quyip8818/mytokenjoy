import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor, fireEvent } from '@testing-library/react'
import { EvaluationsPage } from '@/features/evaluations'
import { renderWithProviders, createMockApis } from '@tests/utils'
import { useSession } from '@/features/session'

const mockWeights = [
  { id: 'w1', dimension: 'quality', weight: 30 },
  { id: 'w2', dimension: 'performance', weight: 20 },
  { id: 'w3', dimension: 'price', weight: 20 },
  { id: 'w4', dimension: 'service', weight: 20 },
  { id: 'w5', dimension: 'compliance', weight: 10 },
]

const mockEvaluationsResponse = {
  items: [
    {
      id: 'e1',
      supplierId: 's1',
      evaluatorId: 'u1',
      period: '2024-Q1',
      quality: 90,
      performance: 85,
      price: 80,
      service: 88,
      compliance: 92,
      totalScore: 87,
      grade: 'B',
      createdAt: '2024-04-01',
      supplierName: 'OpenAI',
      evaluatorName: '管理员',
    },
    {
      id: 'e2',
      supplierId: 's2',
      evaluatorId: 'u1',
      period: '2024-Q1',
      quality: 95,
      performance: 92,
      price: 90,
      service: 94,
      compliance: 96,
      totalScore: 93,
      grade: 'A',
      createdAt: '2024-04-01',
      supplierName: '智谱AI',
      evaluatorName: '管理员',
    },
  ],
  total: 2,
  page: 1,
  pageSize: 10,
}

describe('EvaluationsPage', () => {
  beforeEach(() => {
    useSession.setState({ user: { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' } })
  })

  it('renders evaluation list from API', async () => {
    const apis = createMockApis({
      evaluationsApi: {
        list: vi.fn().mockResolvedValue(mockEvaluationsResponse),
        getWeights: vi.fn().mockResolvedValue(mockWeights),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<EvaluationsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('OpenAI')).toBeInTheDocument()
    })
    expect(screen.getByText('智谱AI')).toBeInTheDocument()
    expect(screen.getAllByText('2024-Q1').length).toBe(2)
  })

  it('displays grade badges', async () => {
    const apis = createMockApis({
      evaluationsApi: {
        list: vi.fn().mockResolvedValue(mockEvaluationsResponse),
        getWeights: vi.fn().mockResolvedValue(mockWeights),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<EvaluationsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('A')).toBeInTheDocument()
      expect(screen.getByText('B')).toBeInTheDocument()
    })
  })

  it('shows 新建评估 button for admin', async () => {
    const apis = createMockApis({
      evaluationsApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
        getWeights: vi.fn().mockResolvedValue(mockWeights),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<EvaluationsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('新建评估')).toBeInTheDocument()
    })
  })

  it('hides 新建评估 button for viewer', async () => {
    useSession.setState({ user: { id: 'u2', username: 'viewer', realName: '查看者', role: 'viewer' } })
    const apis = createMockApis({
      evaluationsApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
        getWeights: vi.fn().mockResolvedValue(mockWeights),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<EvaluationsPage />, { apis })

    await waitFor(() => {
      expect(screen.queryByText('新建评估')).not.toBeInTheDocument()
    })
  })

  it('filters by period on input change', async () => {
    const listFn = vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
    const apis = createMockApis({
      evaluationsApi: {
        list: listFn,
        getWeights: vi.fn().mockResolvedValue(mockWeights),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<EvaluationsPage />, { apis })

    await waitFor(() => expect(listFn).toHaveBeenCalled())

    const input = screen.getByPlaceholderText('评估周期 如 2026-Q3')
    fireEvent.change(input, { target: { value: '2024-Q2' } })

    await waitFor(() => {
      expect(listFn).toHaveBeenCalledWith(
        expect.objectContaining({ period: '2024-Q2', page: 1 }),
      )
    })
  })

  it('shows empty state when no evaluations', async () => {
    const apis = createMockApis({
      evaluationsApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
        getWeights: vi.fn().mockResolvedValue(mockWeights),
      },
      suppliersApi: { options: vi.fn().mockResolvedValue([]) },
    })

    renderWithProviders(<EvaluationsPage />, { apis })

    await waitFor(() => {
      expect(screen.getByText('暂无数据')).toBeInTheDocument()
    })
  })
})

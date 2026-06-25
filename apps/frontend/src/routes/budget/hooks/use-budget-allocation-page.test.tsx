import { waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { BudgetGroup, BudgetNode } from '@/api/types'
import { useBudgetAllocationPage } from './use-budget-allocation-page'
import { createMockApis, renderHookWithProviders } from '@/test-utils'

const mockGroups: BudgetGroup[] = [
  {
    id: 'bg1',
    name: '项目 A',
    budget: 10000,
    consumed: 2000,
    memberIds: ['m1'],
    departmentIds: ['d1'],
  },
]

const mockTree: BudgetNode[] = [
  {
    id: 'n1',
    name: '总部',
    parentId: null,
    budget: 50000,
    consumed: 10000,
    period: '2026-01',
    children: [],
  },
]

describe('useBudgetAllocationPage', () => {
  it('loads budget groups on mount', async () => {
    const apis = createMockApis({
      budgetApi: {
        ...createMockApis().budgetApi,
        getGroups: vi.fn().mockResolvedValue(mockGroups),
        getTree: vi.fn().mockResolvedValue(mockTree),
      },
    })

    const { result } = renderHookWithProviders(() => useBudgetAllocationPage(apis), { apis })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(apis.budgetApi.getGroups).toHaveBeenCalled()
    expect(apis.budgetApi.getTree).toHaveBeenCalled()
    expect(result.current.groups).toEqual(mockGroups)
  })
})

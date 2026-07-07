import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useBudgetPage } from '@/features/budget/hooks/use-budget-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockBudgetGroups, mockBudgetTree } from '@tests/fixtures/budget'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useBudgetPage', () => {
  it('loads budget tree and groups on mount', async () => {
    const apis = createMockApis({
      budgetApi: {
        getTree: vi.fn().mockResolvedValue(mockBudgetTree),
        getGroups: vi.fn().mockResolvedValue(mockBudgetGroups),
        getApprovals: vi.fn().mockResolvedValue([]),
        getOverrunPolicy: vi.fn().mockResolvedValue({
          thresholds: [80],
          notifyEmail: false,
          notifyPhone: false,
          notifyIm: false,
          blockMessage: '',
        }),
      },
    })

    const { result } = renderHookWithProviders(() => useBudgetPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.tree).toEqual(mockBudgetTree)
    })

    expect(apis.budgetApi.getTree).toHaveBeenCalled()
    expect(apis.budgetApi.getGroups).toHaveBeenCalled()
    expect(result.current.selectedTeamId).toBe('n1')
    expect(result.current.pendingCount).toBe(0)
  })
})

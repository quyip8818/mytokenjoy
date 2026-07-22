import { describe, expect, it, vi } from 'vitest'
import { act, waitFor } from '@testing-library/react'
import { useBudgetPage } from '@/features/budget/hooks/use-budget-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockProjects, mockBudgetTree } from '@tests/fixtures/budget'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useBudgetPage', () => {
  it('loads budget tree and groups on mount', async () => {
    const apis = createMockApis({
      budgetApi: {
        getTree: vi.fn().mockResolvedValue(mockBudgetTree),
        getProjects: vi.fn().mockResolvedValue(mockProjects),
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
    expect(apis.budgetApi.getProjects).toHaveBeenCalled()
    expect(result.current.selectedTeamId).toBe('n1')
  })

  it('updates project memberBudgets via budget API', async () => {
    const updateProject = vi.fn().mockResolvedValue(mockProjects[0])
    const apis = createMockApis({
      budgetApi: {
        getTree: vi.fn().mockResolvedValue(mockBudgetTree),
        getProjects: vi.fn().mockResolvedValue(mockProjects),
        getOverrunPolicy: vi.fn().mockResolvedValue({
          thresholds: [80],
          notifyEmail: false,
          notifyPhone: false,
          notifyIm: false,
          blockMessage: '',
        }),
        updateProject,
      },
    })

    const { result } = renderHookWithProviders(() => useBudgetPage(apis), { apis })
    await waitForLoaded(result, 'loading')

    await act(async () => {
      await result.current.updateProject('proj-1', { memberBudgets: { m1: 4000 } })
    })

    expect(updateProject).toHaveBeenCalledWith('proj-1', { memberBudgets: { m1: 4000 } })
  })
})

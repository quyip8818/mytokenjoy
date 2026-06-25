import { describe, expect, it, vi } from 'vitest'
import { useBudgetAllocationPage } from '@/routes/budget/hooks/use-budget-allocation-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockBudgetGroups, mockBudgetTree } from '@tests/fixtures/budget'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useBudgetAllocationPage', () => {
  it('loads budget groups on mount', async () => {
    const apis = createMockApis({
      budgetApi: {
        getGroups: vi.fn().mockResolvedValue(mockBudgetGroups),
        getTree: vi.fn().mockResolvedValue(mockBudgetTree),
      },
    })

    const { result } = renderHookWithProviders(() => useBudgetAllocationPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.budgetApi.getGroups).toHaveBeenCalled()
    expect(apis.budgetApi.getTree).toHaveBeenCalled()
    expect(result.current.groups).toEqual(mockBudgetGroups)
  })
})

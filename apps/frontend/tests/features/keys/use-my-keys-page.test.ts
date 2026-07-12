import { describe, expect, it, vi } from 'vitest'
import { act, waitFor } from '@testing-library/react'
import { useMyKeysPage } from '@/features/keys/hooks/use-my-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useMyKeysPage', () => {
  it('loads member keys and budget summary on mount', async () => {
    const items = [
      { id: 'pk-1', name: 'My Key', status: 'active', scope: 'member' },
      { id: 'pk-2', name: 'Project Key', status: 'active', scope: 'project' },
      { id: 'pk-3', name: 'PM Key', status: 'active', scope: 'project_member' },
    ]
    const budgetSummary = { remaining: 1000, consumed: 200, totalBudget: 1200, reservedPool: 0 }
    const apis = createMockApis({
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items, total: items.length }),
        getBudgetSummary: vi.fn().mockResolvedValue(budgetSummary),
      },
    })

    const { result } = renderHookWithProviders(() => useMyKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual([items[0], items[2]])
    })

    expect(apis.platformKeyApi.list).toHaveBeenCalled()
    expect(apis.platformKeyApi.getBudgetSummary).toHaveBeenCalled()
    expect(result.current.budgetSummary).toEqual(budgetSummary)
  })

  it('opens create workflow with member scope', async () => {
    const budgetSummary = { remaining: 1000, consumed: 200, totalBudget: 1200, reservedPool: 0 }
    const apis = createMockApis({
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
        getBudgetSummary: vi.fn().mockResolvedValue(budgetSummary),
      },
    })

    const { result } = renderHookWithProviders(() => useMyKeysPage(apis), { apis })
    await waitForLoaded(result, 'loading')

    await act(async () => {
      result.current.openCreateKey()
    })

    expect(result.current.budgetSummary?.remaining).toBeGreaterThan(0)
  })
})

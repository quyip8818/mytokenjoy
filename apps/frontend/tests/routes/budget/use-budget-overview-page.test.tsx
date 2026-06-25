import { describe, expect, it, vi } from 'vitest'
import type { BudgetNode } from '@/api/types'
import { useBudgetOverviewPage } from '@/routes/budget/hooks/use-budget-overview-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

const tree: BudgetNode[] = [
  {
    id: 'dept-1',
    name: '总公司',
    parentId: null,
    budget: 100000,
    consumed: 50000,
    period: '2026-06',
    children: [],
  },
]

describe('useBudgetOverviewPage', () => {
  it('exposes period label from API tree', async () => {
    const apis = createMockApis({
      budgetApi: {
        getTree: vi.fn().mockResolvedValue(tree),
      },
    })

    const { result } = renderHookWithProviders(() => useBudgetOverviewPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.periodLabel).toBe('2026 年 6 月')
  })
})

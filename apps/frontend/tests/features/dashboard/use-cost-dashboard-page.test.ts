import { describe, expect, it, vi } from 'vitest'
import { useCostDashboardPage } from '@/features/dashboard/hooks/use-cost-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useCostDashboardPage', () => {
  it('loads cost summary and builds stats on mount', async () => {
    const summary = {
      totalCost: 1000,
      totalCostMom: 5,
      totalTokens: 2000000,
      totalRequests: 100,
      avgCostPerRequest: 10,
      avgCostPerRequestMom: 0,
      avgCostPerMember: 500,
      avgCostPerMemberMom: 0,
      totalRequestsMom: 0,
    }
    const apis = createMockApis({
      dashboardApi: {
        getCostSummary: vi.fn().mockResolvedValue(summary),
        getDailyCosts: vi.fn().mockResolvedValue([]),
        getDepartmentCosts: vi.fn().mockResolvedValue([]),
        getTopConsumers: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useCostDashboardPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.dashboardApi.getCostSummary).toHaveBeenCalled()
    expect(result.current.summary).toEqual(summary)
    expect(result.current.stats).toHaveLength(5)
    expect(result.current.stats[0]?.label).toBe('总花费')
  })
})

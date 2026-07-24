import { describe, expect, it, vi } from 'vitest'
import { useMyDashboardPage } from '@/features/mydashboard/hooks/use-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useMyDashboardPage', () => {
  it('returns dashboard stats on successful load', async () => {
    const dashboard = {
      account: { budgetRemaining: 500, totalSpent: 200 },
      usageStats: { requestCount: 100, totalCount: 150 },
      resourceConsumption: { totalCost: 80, totalTokens: 1000 },
      performance: { avgRPM: 10, avgTPM: 500 },
      consumptionTrend: [],
      consumptionDistribution: [],
      callDistribution: [],
      callRanking: [],
    }
    const apis = createMockApis({
      meApi: { getDashboard: vi.fn().mockResolvedValue(dashboard) },
    })

    const { result } = renderHookWithProviders(() => useMyDashboardPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.meApi.getDashboard).toHaveBeenCalled()
    expect(result.current.accountData.budgetRemaining).toBe(500)
    expect(result.current.usageStats.requestCount).toBe(100)
  })
})

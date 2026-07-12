import { describe, expect, it, vi } from 'vitest'
import { useMemberDashboardPage } from '@/features/member/hooks/use-member-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useMemberDashboardPage', () => {
  it('loads member dashboard on mount', async () => {
    const dashboard = {
      account: { budgetRemaining: 100, totalSpent: 50 },
      usageStats: { requestCount: 10, totalCount: 10 },
      resourceConsumption: { totalCost: 20, totalTokens: 1000 },
      performance: { avgRPM: 1.5, avgTPM: 200 },
      consumptionTrend: [],
      consumptionDistribution: [],
      callDistribution: [],
      callRanking: [],
    }
    const apis = createMockApis({
      meApi: {
        getDashboard: vi.fn().mockResolvedValue(dashboard),
      },
    })

    const { result } = renderHookWithProviders(() => useMemberDashboardPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.meApi.getDashboard).toHaveBeenCalled()
    expect(result.current.accountData).toEqual(dashboard.account)
  })
})

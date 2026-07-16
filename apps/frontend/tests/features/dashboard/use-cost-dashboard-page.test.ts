import { describe, expect, it, vi } from 'vitest'
import { useCostDashboardPage } from '@/features/dashboard/hooks/use-cost-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { createDashboardApiMock } from '@tests/helpers/factories'
import { mockCostSummary } from '@tests/fixtures/dashboard'

describe('useCostDashboardPage', () => {
  it('loads cost summary and builds stats on mount', async () => {
    const apis = createMockApis({
      dashboardApi: createDashboardApiMock({
        getCostSummary: vi.fn().mockResolvedValue(mockCostSummary),
      }),
    })

    const { result } = renderHookWithProviders(
      () => useCostDashboardPage({ deptId: null, injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')

    expect(apis.dashboardApi.getCostSummary).toHaveBeenCalled()
    expect(result.current.summary).toEqual(mockCostSummary)
    expect(result.current.stats).toHaveLength(3)
    expect(result.current.stats[0]?.label).toBe('总花费')
  })

  it('passes deptId as parentId to getDepartmentCosts', async () => {
    const apis = createMockApis({
      dashboardApi: createDashboardApiMock(),
    })

    const { result } = renderHookWithProviders(
      () => useCostDashboardPage({ deptId: 'd1', injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')

    expect(apis.dashboardApi.getDepartmentCosts).toHaveBeenCalledWith(
      expect.objectContaining({ parentId: 'd1' }),
    )
  })
})

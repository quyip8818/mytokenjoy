import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useUsageDashboardPage } from '@/features/dashboard/hooks/use-usage-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { createDashboardApiMock } from '@tests/helpers/factories'

describe('useUsageDashboardPage', () => {
  it('loads department and model usage on mount', async () => {
    const apis = createMockApis({
      dashboardApi: createDashboardApiMock({
        getDepartmentUsage: vi.fn().mockResolvedValue([
          {
            departmentId: 'd1',
            departmentName: 'HQ',
            budget: 1000,
            consumed: 500,
            memberCount: 5,
            topModel: 'gpt-4',
          },
        ]),
        getModelUsage: vi.fn().mockResolvedValue([
          {
            callType: 'gpt-4',
            modelName: 'GPT-4',
            tokens: 50,
            cost: 1,
            requests: 1,
            percentage: 100,
            provider: 'openai',
          },
        ]),
      }),
    })

    const { result } = renderHookWithProviders(
      () => useUsageDashboardPage({ deptId: null, injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.departmentUsage).toHaveLength(1)
    })

    expect(apis.dashboardApi.getDepartmentUsage).toHaveBeenCalled()
    expect(apis.dashboardApi.getModelUsage).toHaveBeenCalled()
  })

  it('passes period params when deptId changes', async () => {
    const apis = createMockApis({
      dashboardApi: createDashboardApiMock(),
    })

    const { result } = renderHookWithProviders(
      () => useUsageDashboardPage({ deptId: 'd1', injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    expect(apis.dashboardApi.getDepartmentUsage).toHaveBeenCalledWith(
      expect.objectContaining({ period: 'current_month' }),
    )
  })
})

import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useUsageDashboardPage } from '@/features/dashboard/hooks/use-usage-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useUsageDashboardPage', () => {
  it('loads team and model usage on mount', async () => {
    const apis = createMockApis({
      dashboardApi: {
        getTeamUsage: vi
          .fn()
          .mockResolvedValue([{ departmentId: 'd1', departmentName: 'HQ', quota: 1000, consumed: 500, memberCount: 5, topModel: 'gpt-4' }]),
        getModelUsage: vi.fn().mockResolvedValue([
          { callType: 'gpt-4', modelName: 'GPT-4', tokens: 50, cost: 1, requests: 1, percentage: 100, provider: 'openai' },
        ]),
      },
    })

    const { result } = renderHookWithProviders(
      () => useUsageDashboardPage({ deptId: null, injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.teamUsage).toHaveLength(1)
    })

    expect(apis.dashboardApi.getTeamUsage).toHaveBeenCalled()
    expect(apis.dashboardApi.getModelUsage).toHaveBeenCalled()
  })

  it('passes period params when deptId changes', async () => {
    const apis = createMockApis({
      dashboardApi: {
        getTeamUsage: vi.fn().mockResolvedValue([]),
        getModelUsage: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(
      () => useUsageDashboardPage({ deptId: 'd1', injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    expect(apis.dashboardApi.getTeamUsage).toHaveBeenCalledWith(
      expect.objectContaining({ period: 'current_month' }),
    )
  })
})

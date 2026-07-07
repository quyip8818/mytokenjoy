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
          .mockResolvedValue([{ departmentId: 'd1', departmentName: 'HQ', tokens: 100 }]),
        getModelUsage: vi.fn().mockResolvedValue([{ model: 'gpt-4', tokens: 50 }]),
      },
    })

    const { result } = renderHookWithProviders(() => useUsageDashboardPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.teamUsage).toHaveLength(1)
    })

    expect(apis.dashboardApi.getTeamUsage).toHaveBeenCalled()
    expect(apis.dashboardApi.getModelUsage).toHaveBeenCalled()
    expect(result.current.modelUsage).toHaveLength(1)
  })
})

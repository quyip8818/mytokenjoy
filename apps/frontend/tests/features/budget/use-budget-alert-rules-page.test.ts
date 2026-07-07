import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useBudgetAlertRulesPage } from '@/features/budget/hooks/use-budget-alert-rules-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockBudgetGroups, mockBudgetTree } from '@tests/fixtures/budget'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useBudgetAlertRulesPage', () => {
  it('loads alert rules on mount', async () => {
    const rules = [
      {
        id: 'alert-1',
        nodeId: 'n1',
        nodeName: 'Engineering',
        thresholds: [80],
        notifyRoleIds: ['role-1'],
        enabled: true,
      },
    ]
    const apis = createMockApis({
      budgetApi: {
        getAlerts: vi.fn().mockResolvedValue(rules),
        getGroups: vi.fn().mockResolvedValue(mockBudgetGroups),
        getTree: vi.fn().mockResolvedValue(mockBudgetTree),
      },
      roleApi: {
        list: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useBudgetAlertRulesPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.rules).toHaveLength(1)
    })

    expect(apis.budgetApi.getAlerts).toHaveBeenCalled()
    expect(result.current.projects.length).toBeGreaterThan(0)
  })
})

import { describe, expect, it, vi } from 'vitest'
import { act, waitFor } from '@testing-library/react'
import { useBudgetAlertRulesPage } from '@/features/budget/hooks/use-budget-alert-rules-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockProjects, mockBudgetTree } from '@tests/fixtures/budget'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

const mockRules = [
  {
    id: 'alert-1',
    nodeId: 'dept-1',
    nodeName: '总公司',
    thresholds: [80, 90],
    notifyRoleIds: ['role-1'],
    enabled: true,
  },
  {
    id: 'alert-2',
    nodeId: 'proj-1',
    nodeName: '项目A',
    thresholds: [100],
    notifyRoleIds: ['role-2'],
    enabled: false,
  },
  {
    id: 'alert-3',
    nodeId: 'dept-2',
    nodeName: '技术部',
    thresholds: [90, 100],
    notifyRoleIds: ['role-1', 'role-2'],
    enabled: true,
  },
]

function createAlertApis() {
  return createMockApis({
    budgetApi: {
      getAlerts: vi.fn().mockResolvedValue(mockRules),
      getProjects: vi.fn().mockResolvedValue(mockProjects),
      getTree: vi.fn().mockResolvedValue(mockBudgetTree),
    },
    roleApi: {
      list: vi.fn().mockResolvedValue([
        { id: 'role-1', name: '超级管理员' },
        { id: 'role-2', name: '组织管理员' },
      ]),
    },
  })
}

describe('useBudgetAlertRulesPage', () => {
  it('loads alert rules on mount', async () => {
    const apis = createAlertApis()

    const { result } = renderHookWithProviders(() => useBudgetAlertRulesPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.rules).toHaveLength(3)
    })

    expect(apis.budgetApi.getAlerts).toHaveBeenCalled()
  })

  it('computes stats correctly', async () => {
    const apis = createAlertApis()

    const { result } = renderHookWithProviders(() => useBudgetAlertRulesPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.stats.total).toBe(3)
    })

    expect(result.current.stats.enabled).toBe(2)
    expect(result.current.stats.teamCoverage.covered).toBe(2) // dept-1, dept-2
    expect(result.current.stats.projectCoverage.covered).toBe(1) // proj-1
  })

  it('filters rules by type', async () => {
    const apis = createAlertApis()

    const { result } = renderHookWithProviders(() => useBudgetAlertRulesPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.allRules).toHaveLength(3)
    })

    act(() => {
      result.current.setTypeFilter('team')
    })

    expect(result.current.rules.every((r) => r.targetType === 'team')).toBe(true)
    expect(result.current.rules).toHaveLength(2)

    act(() => {
      result.current.setTypeFilter('project')
    })

    expect(result.current.rules.every((r) => r.targetType === 'project')).toBe(true)
    expect(result.current.rules).toHaveLength(1)
  })

  it('filters rules by status', async () => {
    const apis = createAlertApis()

    const { result } = renderHookWithProviders(() => useBudgetAlertRulesPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.allRules).toHaveLength(3)
    })

    act(() => {
      result.current.setStatusFilter('enabled')
    })

    expect(result.current.rules.every((r) => r.enabled)).toBe(true)
    expect(result.current.rules).toHaveLength(2)

    act(() => {
      result.current.setStatusFilter('disabled')
    })

    expect(result.current.rules.every((r) => !r.enabled)).toBe(true)
    expect(result.current.rules).toHaveLength(1)
  })

  it('filters rules by search keyword', async () => {
    const apis = createAlertApis()

    const { result } = renderHookWithProviders(() => useBudgetAlertRulesPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.allRules).toHaveLength(3)
    })

    act(() => {
      result.current.setSearch('总公司')
    })

    expect(result.current.rules).toHaveLength(1)
    expect(result.current.rules[0].targetName).toBe('总公司')
  })
})

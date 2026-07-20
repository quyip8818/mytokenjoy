import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import {
  formatBudgetContext,
  useKeyFormBudget,
} from '@/features/workflow/workflows/key-form/use-key-form-budget'
import { createMockApis, renderHookWithProviders } from '@tests/utils'

describe('formatBudgetContext', () => {
  it('includes remaining budget and department label', () => {
    // remaining=600000 quota → ¥1.20 at DEFAULT_QUOTA_PER_UNIT=500000
    expect(
      formatBudgetContext(
        { remaining: 600000, consumed: 150000, totalBudget: 750000, reservedPool: 0 },
        '研发部',
      ),
    ).toContain('研发部')
    expect(
      formatBudgetContext(
        { remaining: 600000, consumed: 150000, totalBudget: 750000, reservedPool: 0 },
        '研发部',
      ),
    ).toContain('¥1.2')
  })
})

describe('useKeyFormBudget', () => {
  it('loads project_member sub-budget remaining from memberBudgets minus allocated keys', async () => {
    const apis = createMockApis({
      budgetApi: {
        getProjects: vi.fn().mockResolvedValue([
          {
            id: 'proj-1',
            name: '项目 A',
            budget: 10000,
            consumed: 1000,
            memberIds: ['m1'],
            memberBudgets: { m1: 5000 },
            ownerDepartmentId: 'd1',
          },
        ]),
      },
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({
          items: [
            { id: 'pk-1', status: 'active', memberId: 'm1', budget: 2000, scope: 'project_member' },
          ],
          total: 1,
        }),
      },
    })

    const { result } = renderHookWithProviders(
      () =>
        useKeyFormBudget({
          isCreate: true,
          scope: 'project_member',
          effectiveMemberId: 'm1',
          projectId: 'proj-1',
          budget: '1',
          adminCreate: true,
          injectedApis: apis,
        }),
      { apis },
    )

    await waitFor(() => {
      expect(result.current.subBudgetRemaining).toBe(3000)
    })
    expect(apis.platformKeyApi.list).toHaveBeenCalledWith({
      projectId: 'proj-1',
      scope: 'project_member',
      memberId: 'm1',
    })
    expect(result.current.subBudgetExceeds).toBe(false)
  })

  it('loads project budget remaining from all active keys on the project', async () => {
    const apis = createMockApis({
      budgetApi: {
        getProjects: vi.fn().mockResolvedValue([
          {
            id: 'proj-1',
            name: '项目 A',
            budget: 10000,
            consumed: 1000,
            memberIds: ['m1'],
            memberBudgets: { m1: 5000 },
            ownerDepartmentId: 'd1',
          },
        ]),
      },
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({
          items: [
            { id: 'pk-1', status: 'active', budget: 2000, scope: 'project' },
            { id: 'pk-2', status: 'active', memberId: 'm1', budget: 1500, scope: 'project_member' },
            { id: 'pk-3', status: 'disabled', budget: 999, scope: 'project_member' },
          ],
          total: 3,
        }),
      },
    })

    const { result } = renderHookWithProviders(
      () =>
        useKeyFormBudget({
          isCreate: true,
          scope: 'project',
          effectiveMemberId: '',
          projectId: 'proj-1',
          budget: '1',
          adminCreate: true,
          injectedApis: apis,
        }),
      { apis },
    )

    await waitFor(() => {
      expect(result.current.projectBudgetRemaining).toBe(5500)
    })
    expect(apis.platformKeyApi.list).toHaveBeenCalledWith({ projectId: 'proj-1' })
  })
})

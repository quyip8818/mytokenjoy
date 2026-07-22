import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useApprovalPendingCountQuery } from '@/features/org/hooks/use-approval-pending-count-query'
import { PERMISSION } from '@/lib/permissions'
import { createMockApis, renderHookWithProviders } from '@tests/utils'

describe('useApprovalPendingCountQuery', () => {
  it('loads pending approval count when user can approve', async () => {
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 2 }),
      },
    })

    const { result } = renderHookWithProviders(
      () => useApprovalPendingCountQuery({ injectedApis: apis }),
      {
        apis,
        permissions: [PERMISSION.BUDGET_APPROVE],
      },
    )

    await waitFor(() => {
      expect(result.current.data).toBe(2)
    })

    expect(apis.approvalApi.list).toHaveBeenCalledWith({ status: 'pending', limit: 0 })
  })

  it('skips fetch when user cannot approve', async () => {
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
    })

    const { result } = renderHookWithProviders(
      () => useApprovalPendingCountQuery({ injectedApis: apis }),
      {
        apis,
        permissions: [],
      },
    )

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toBeUndefined()
    expect(apis.approvalApi.list).not.toHaveBeenCalled()
  })
})

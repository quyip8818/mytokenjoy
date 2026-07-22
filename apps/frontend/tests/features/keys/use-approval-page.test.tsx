import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useApprovalPage } from '@/features/approval/hooks/use-approval-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockApprovals } from '@tests/fixtures/approvals'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useApprovalPage', () => {
  it('loads pending approvals on mount', async () => {
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue({ items: mockApprovals, total: 1 }),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.approvalApi.list).toHaveBeenCalledWith({ status: 'pending' })
    expect(result.current.approvals).toEqual(mockApprovals)
    expect(result.current.pendingCount).toBe(1)
  })

  it('switches tab filter', async () => {
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setTab('approved')
    })

    await waitForLoaded(result, 'loading')

    expect(apis.approvalApi.list).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'approved' }),
    )
  })
})

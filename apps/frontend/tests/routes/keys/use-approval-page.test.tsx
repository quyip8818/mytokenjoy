import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useApprovalPage } from '@/routes/keys/hooks/use-approval-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockApprovals } from '@tests/fixtures/approvals'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useApprovalPage', () => {
  it('loads pending approvals on mount', async () => {
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue(mockApprovals),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.approvalApi.list).toHaveBeenCalledWith({ tab: 'pending' })
    expect(result.current.approvals).toEqual(mockApprovals)
    expect(result.current.pendingCount).toBe(1)
    expect(result.current.hasKeyType).toBe(true)
    expect(result.current.hasQuotaType).toBe(false)
  })

  it('switches tab filter', async () => {
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setTab('mine')
    })

    await waitForLoaded(result, 'loading')

    expect(apis.approvalApi.list).toHaveBeenCalledWith(expect.objectContaining({ tab: 'mine' }))
  })
})

import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useApprovalPage } from '@/features/keys/hooks/use-approval-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useApprovalPage', () => {
  it('loads approvals and computes pending count on mount', async () => {
    const approvals = [
      {
        id: 'a1',
        type: 'key' as const,
        status: 'pending' as const,
        memberId: 'm1',
        memberName: '张三',
      },
      {
        id: 'a2',
        type: 'quota' as const,
        status: 'approved' as const,
        memberId: 'm2',
        memberName: '李四',
      },
    ]
    const apis = createMockApis({
      approvalApi: {
        list: vi.fn().mockResolvedValue(approvals),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.approvals).toEqual(approvals)
    })

    expect(apis.approvalApi.list).toHaveBeenCalled()
    expect(result.current.pendingCount).toBe(1)
    expect(result.current.hasKeyType).toBe(true)
    expect(result.current.hasQuotaType).toBe(true)
  })
})

import { act, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { KeyApproval } from '@/api/types'
import { useApprovalPage } from './use-approval-page'
import { createMockApis, renderHookWithProviders } from '@/test-utils'

const mockApprovals: KeyApproval[] = [
  {
    id: 'a1',
    type: 'key',
    applicant: '张三',
    applicantId: 'm1',
    department: '研发部',
    reason: '需要 API 访问',
    requestedQuota: 0,
    requestedModels: ['gpt-4'],
    status: 'pending',
    approver: null,
    createdAt: '2026-01-01',
    resolvedAt: null,
  },
]

describe('useApprovalPage', () => {
  it('loads pending approvals on mount', async () => {
    const apis = createMockApis({
      approvalApi: {
        ...createMockApis().approvalApi,
        list: vi.fn().mockResolvedValue(mockApprovals),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(apis.approvalApi.list).toHaveBeenCalledWith({ tab: 'pending' })
    expect(result.current.approvals).toEqual(mockApprovals)
    expect(result.current.pendingCount).toBe(1)
    expect(result.current.hasKeyType).toBe(true)
    expect(result.current.hasQuotaType).toBe(false)
  })

  it('switches tab filter', async () => {
    const apis = createMockApis({
      approvalApi: {
        ...createMockApis().approvalApi,
        list: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useApprovalPage(apis), { apis })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    act(() => {
      result.current.setTab('mine')
    })

    await waitFor(() => {
      expect(apis.approvalApi.list).toHaveBeenCalledWith(expect.objectContaining({ tab: 'mine' }))
    })
  })
})

import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useAuditMemberOptions } from '@/features/audit/hooks/use-audit-member-options'
import { createMockApis, renderHookWithProviders } from '@tests/utils'

describe('useAuditMemberOptions', () => {
  it('loads member options on mount', async () => {
    const members = [
      {
        id: 'm1',
        companyId: 1,
        name: '张三',
        phone: '',
        email: 'zhangsan@example.com',
        departmentId: 'd1',
        departmentName: '总部',
        status: 'active' as const,
        roles: [],
        source: 'manual' as const,
      },
    ]
    const apis = createMockApis({
      memberApi: {
        list: vi.fn().mockResolvedValue({ items: members, total: 1, page: 1, pageSize: 100 }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditMemberOptions(apis), { apis })

    await waitFor(() => {
      expect(result.current.members).toEqual(members)
    })

    expect(apis.memberApi.list).toHaveBeenCalled()
  })
})

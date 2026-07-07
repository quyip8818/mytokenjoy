import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useAuditOperationsPage } from '@/features/audit/hooks/use-audit-operations-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useAuditOperationsPage', () => {
  it('loads operation logs and member options on mount', async () => {
    const logs = [{ id: 'op-1', action: 'member_add' }]
    const members = [
      {
        id: 'm1',
        companyId: 1,
        name: '李四',
        phone: '',
        email: 'lisi@example.com',
        departmentId: 'd1',
        departmentName: '总部',
        status: 'active' as const,
        roles: [],
        source: 'manual' as const,
      },
    ]
    const apis = createMockApis({
      auditApi: {
        getOperations: vi.fn().mockResolvedValue({ items: logs, total: 1, page: 1, pageSize: 20 }),
      },
      memberApi: {
        list: vi.fn().mockResolvedValue({ items: members, total: 1, page: 1, pageSize: 100 }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditOperationsPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.logs).toEqual(logs)
    })

    expect(apis.auditApi.getOperations).toHaveBeenCalled()
    expect(result.current.memberOptions).toEqual({ m1: '李四' })
  })
})

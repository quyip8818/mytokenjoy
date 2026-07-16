import { act, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AUDIT_PAGE_SIZE } from '@/features/audit/lib/constants'
import { useAuditOperationsPage } from '@/features/audit/hooks/use-audit-operations-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { createAuditApiMock, createPaginatedResponse } from '@tests/helpers/factories'
import { mockOperationLogs } from '@tests/fixtures/call-logs'
import { mockMembers } from '@tests/fixtures/members'

describe('useAuditOperationsPage', () => {
  it('loads operation logs and member options on mount', async () => {
    const apis = createMockApis({
      auditApi: createAuditApiMock({
        getOperations: vi.fn().mockResolvedValue(createPaginatedResponse(mockOperationLogs)),
      }),
      memberApi: {
        list: vi.fn().mockResolvedValue(createPaginatedResponse(mockMembers, { pageSize: 100 })),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditOperationsPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.logs).toEqual(mockOperationLogs)
    })

    expect(apis.auditApi.getOperations).toHaveBeenCalled()
    expect(result.current.memberOptions).toEqual({ m1: '张三', m2: '李四' })
  })

  it('loads operation logs with pagination params', async () => {
    const apis = createMockApis({
      auditApi: createAuditApiMock({
        getOperations: vi.fn().mockResolvedValue(
          createPaginatedResponse([], { pageSize: AUDIT_PAGE_SIZE }),
        ),
      }),
    })

    const { result } = renderHookWithProviders(() => useAuditOperationsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getOperations).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: AUDIT_PAGE_SIZE }),
    )
  })

  it('requests the selected page', async () => {
    const apis = createMockApis({
      auditApi: createAuditApiMock({
        getOperations: vi.fn().mockResolvedValue(
          createPaginatedResponse([], { total: 40, pageSize: AUDIT_PAGE_SIZE }),
        ),
      }),
    })

    const { result } = renderHookWithProviders(() => useAuditOperationsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setPage(2)
    })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getOperations).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 2, pageSize: AUDIT_PAGE_SIZE }),
    )
  })
})

import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AUDIT_PAGE_SIZE } from '@/features/audit/lib/constants'
import { useAuditOperationsPage } from '@/features/audit/hooks/use-audit-operations-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useAuditOperationsPage', () => {
  it('loads operation logs with pagination params', async () => {
    const apis = createMockApis({
      auditApi: {
        getOperations: vi.fn().mockResolvedValue({
          items: [],
          total: 0,
          page: 1,
          pageSize: AUDIT_PAGE_SIZE,
        }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditOperationsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getOperations).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: AUDIT_PAGE_SIZE }),
    )
  })

  it('requests the selected page', async () => {
    const apis = createMockApis({
      auditApi: {
        getOperations: vi.fn().mockResolvedValue({
          items: [],
          total: 40,
          page: 1,
          pageSize: AUDIT_PAGE_SIZE,
        }),
      },
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

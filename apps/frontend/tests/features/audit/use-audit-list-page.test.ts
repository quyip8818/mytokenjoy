import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useAuditListPage } from '@/features/audit/hooks/use-audit-list-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useAuditListPage', () => {
  it('loads paginated audit items on mount', async () => {
    const items = [{ id: 'log-1', action: 'member_add' }]
    const apis = createMockApis({
      auditApi: {
        getOperations: vi.fn().mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 }),
      },
    })

    const { result } = renderHookWithProviders(
      () =>
        useAuditListPage({
          initialFilter: { keyword: '' },
          toQueryParams: (filter) => ({ keyword: filter.keyword }),
          fetchPage: (a, query) => a.auditApi.getOperations(query),
          injectedApis: apis,
          queryKeyFactory: ({ filter, page }) => ['audit-test', filter, page],
        }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.items).toEqual(items)
    })

    expect(apis.auditApi.getOperations).toHaveBeenCalled()
    expect(result.current.total).toBe(1)
  })
})

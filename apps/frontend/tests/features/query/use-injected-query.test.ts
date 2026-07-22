import { describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import { createMockApis } from '@tests/utils'
import { createTestWrapper } from '@tests/utils'

describe('useInjectedQuery', () => {
  it('loads data on mount and exposes refresh', async () => {
    const apis = createMockApis({
      sessionApi: {
        getCurrent: vi.fn().mockResolvedValue({
          member: {
            id: 'm1',
            alias: 'Test',
            avatar: '',
            departmentId: 'd1',
            departmentName: 'Dept',
            status: 'active',
            roles: [],
            source: 'manual',
          },
          user: { name: 'Test' },
          permissions: [],
          readOnly: false,
        }),
      },
    })

    const { result } = renderHook(
      () =>
        useInjectedQuery({
          injectedApis: apis,
          queryKey: ['session', 'current'],
          queryFn: (a) => a.sessionApi.getCurrent(),
        }),
      { wrapper: createTestWrapper({ apis }) },
    )

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data?.member.id).toBe('m1')
    expect(result.current.error).toBeNull()
  })
})

import { describe, expect, it, vi } from 'vitest'
import { useMemberCallLogsPage } from '@/features/member/hooks/use-member-call-logs-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useMemberCallLogsPage', () => {
  it('loads call logs for current member', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
    })

    const { result } = renderHookWithProviders(() => useMemberCallLogsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenCalledWith(
      expect.objectContaining({ callerId: 'm-admin', page: 1, pageSize: 20 }),
    )
    expect(result.current.logs).toEqual([])
  })
})

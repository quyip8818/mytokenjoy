import { describe, expect, it, vi } from 'vitest'
import { useMyCallLogsPage } from '@/features/mydashboard/hooks/use-call-logs-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useMyCallLogsPage', () => {
  it('returns call logs for current member', async () => {
    const callLogs = { items: [{ id: '1', model: 'gpt-4' }], total: 1 }
    const apis = createMockApis({
      auditApi: { getCalls: vi.fn().mockResolvedValue(callLogs) },
    })

    const { result } = renderHookWithProviders(() => useMyCallLogsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.logs).toHaveLength(1)
    expect(result.current.total).toBe(1)
  })
})

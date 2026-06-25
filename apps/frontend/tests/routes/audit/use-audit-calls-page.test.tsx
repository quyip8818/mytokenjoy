import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useAuditCallsPage } from '@/routes/audit/hooks/use-audit-calls-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import type { CallLog } from '@/api/types'

const mockLogs: CallLog[] = [
  {
    id: 'cl-1',
    caller: '张三',
    callerId: 'm-1',
    callerType: 'member',
    model: 'gpt-4o',
    provider: 'openai',
    inputTokens: 100,
    outputTokens: 50,
    latencyMs: 200,
    status: 'success',
    cost: 1.5,
    createdAt: '2026-06-19 10:00',
    inputPreview: 'hello',
    outputPreview: 'world',
  },
]

describe('useAuditCallsPage', () => {
  it('loads call logs on mount', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue({ items: mockLogs, total: 1 }),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenCalled()
    expect(result.current.logs).toEqual(mockLogs)
  })

  it('applies status filter', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue({ items: [], total: 0 }),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setStatusFilter('error')
    })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'error' }),
    )
  })
})

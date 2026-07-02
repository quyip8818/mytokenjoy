import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AUDIT_PAGE_SIZE } from '@/lib/audit-constants'
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
    previewSnippet: 'hello',
  },
]

const paginatedResponse = {
  items: mockLogs,
  total: 25,
  page: 1,
  pageSize: AUDIT_PAGE_SIZE,
}

describe('useAuditCallsPage', () => {
  it('loads call logs on mount with pagination params', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue(paginatedResponse),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: AUDIT_PAGE_SIZE }),
    )
    expect(result.current.logs).toEqual(mockLogs)
    expect(result.current.total).toBe(25)
    expect(result.current.totalPages).toBe(2)
  })

  it('applies status filter and resets page', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: AUDIT_PAGE_SIZE }),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setPage(2)
    })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setStatusFilter('error')
    })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: 'error', page: 1, pageSize: AUDIT_PAGE_SIZE }),
    )
  })

  it('applies model filter', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: AUDIT_PAGE_SIZE }),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setModelFilter('gpt-4o')
    })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenLastCalledWith(
      expect.objectContaining({ model: 'gpt-4o', page: 1, pageSize: AUDIT_PAGE_SIZE }),
    )
  })

  it('requests the selected page', async () => {
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue(paginatedResponse),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setPage(2)
    })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 2, pageSize: AUDIT_PAGE_SIZE }),
    )
  })
})

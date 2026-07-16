import { act, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AUDIT_PAGE_SIZE } from '@/features/audit/lib/constants'
import { useAuditCallsPage } from '@/features/audit/hooks/use-audit-calls-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { createAuditApiMock, createPaginatedResponse } from '@tests/helpers/factories'
import { mockCallLogs } from '@tests/fixtures/call-logs'
import { mockModelRefs } from '@tests/fixtures/models'
import { mockMembers } from '@tests/fixtures/members'

describe('useAuditCallsPage', () => {
  it('loads call logs and filter options on mount', async () => {
    const apis = createMockApis({
      auditApi: createAuditApiMock({
        getCalls: vi.fn().mockResolvedValue(createPaginatedResponse(mockCallLogs)),
      }),
      modelApi: {
        list: vi.fn().mockResolvedValue(mockModelRefs),
      },
      memberApi: {
        list: vi.fn().mockResolvedValue(createPaginatedResponse(mockMembers, { pageSize: 100 })),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.logs).toEqual(mockCallLogs)
    })

    expect(apis.auditApi.getCalls).toHaveBeenCalled()
    expect(result.current.modelOptions).toEqual({
      'gpt-4': 'GPT-4',
      'gpt-4o': 'GPT-4o',
      'claude-3': 'Claude 3',
    })
    expect(result.current.memberOptions).toEqual({ m1: '张三', m2: '李四' })
    expect(result.current.contentRetentionEnabled).toBe(true)
  })

  it('loads call logs on mount with pagination params', async () => {
    const paginatedResponse = createPaginatedResponse(mockCallLogs, {
      total: 25,
      pageSize: AUDIT_PAGE_SIZE,
    })
    const apis = createMockApis({
      auditApi: createAuditApiMock({
        getCalls: vi.fn().mockResolvedValue(paginatedResponse),
      }),
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getCalls).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: AUDIT_PAGE_SIZE }),
    )
    expect(result.current.logs).toEqual(mockCallLogs)
    expect(result.current.total).toBe(25)
    expect(result.current.totalPages).toBe(2)
  })

  it('applies status filter and resets page', async () => {
    const apis = createMockApis({
      auditApi: createAuditApiMock({
        getCalls: vi
          .fn()
          .mockResolvedValue(createPaginatedResponse([], { pageSize: AUDIT_PAGE_SIZE })),
      }),
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
      auditApi: createAuditApiMock({
        getCalls: vi
          .fn()
          .mockResolvedValue(createPaginatedResponse([], { pageSize: AUDIT_PAGE_SIZE })),
      }),
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
      auditApi: createAuditApiMock({
        getCalls: vi
          .fn()
          .mockResolvedValue(
            createPaginatedResponse(mockCallLogs, { total: 25, pageSize: AUDIT_PAGE_SIZE }),
          ),
      }),
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

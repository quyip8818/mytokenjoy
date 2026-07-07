import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useAuditCallsPage } from '@/features/audit/hooks/use-audit-calls-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useAuditCallsPage', () => {
  it('loads call logs and filter options on mount', async () => {
    const logs = [{ id: 'call-1', status: 'success', model: 'gpt-4' }]
    const models = [
      { id: 'm1', name: 'gpt-4', displayName: 'GPT-4', provider: 'openai', enabled: true },
    ]
    const members = [
      {
        id: 'm1',
        companyId: 1,
        name: '张三',
        phone: '',
        email: 'zhangsan@example.com',
        departmentId: 'd1',
        departmentName: '总部',
        status: 'active' as const,
        roles: [],
        source: 'manual' as const,
      },
    ]
    const apis = createMockApis({
      auditApi: {
        getCalls: vi.fn().mockResolvedValue({ items: logs, total: 1, page: 1, pageSize: 20 }),
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
      },
      modelApi: {
        list: vi.fn().mockResolvedValue(models),
      },
      memberApi: {
        list: vi.fn().mockResolvedValue({ items: members, total: 1, page: 1, pageSize: 100 }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditCallsPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.logs).toEqual(logs)
    })

    expect(apis.auditApi.getCalls).toHaveBeenCalled()
    expect(result.current.modelOptions).toEqual({ 'gpt-4': 'GPT-4' })
    expect(result.current.memberOptions).toEqual({ m1: '张三' })
    expect(result.current.contentRetentionEnabled).toBe(true)
  })
})

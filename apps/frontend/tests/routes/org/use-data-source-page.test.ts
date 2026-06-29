import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { toast } from 'sonner'
import { useDataSourcePage } from '@/routes/org/hooks/use-data-source-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
  },
}))

describe('useDataSourcePage', () => {
  it('loads data source status and sync config', async () => {
    const getStatus = vi.fn().mockResolvedValue({
      connected: true,
      platform: 'feishu',
      lastImport: '2026-06-10T10:00:00Z',
    })
    const getConfig = vi.fn().mockResolvedValue({
      enabled: true,
      frequencyHours: 24,
      startTime: '02:00',
    })
    const apis = createMockApis({
      dataSourceApi: { getStatus },
      syncApi: {
        getConfig,
        getLogs: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
    })
    const { result } = renderHookWithProviders(() => useDataSourcePage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(getStatus).toHaveBeenCalled()
    expect(getConfig).toHaveBeenCalled()
    expect(result.current.status?.connected).toBe(true)
  })

  it('stores import result after handleImport', async () => {
    const importResult = {
      successMembers: 12,
      successDepartments: 3,
      failedMembers: 0,
      failedDepartments: 0,
      failures: [],
    }
    const importFn = vi.fn().mockResolvedValue(importResult)
    const apis = createMockApis({
      dataSourceApi: {
        getStatus: vi.fn().mockResolvedValue({ connected: true, platform: 'feishu' }),
        import: importFn,
      },
      syncApi: {
        getConfig: vi
          .fn()
          .mockResolvedValue({ enabled: false, frequencyHours: 24, startTime: '02:00' }),
        getLogs: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
    })
    const { result } = renderHookWithProviders(() => useDataSourcePage(apis), { apis })

    await waitForLoaded(result, 'loading')

    await act(async () => {
      await result.current.handleImport()
    })

    expect(importFn).toHaveBeenCalled()
    expect(result.current.displayImportResult).toEqual(importResult)
    expect(toast.success).toHaveBeenCalledWith('导入完成：12 人 / 3 个部门')
  })

  it('triggers manual sync and refreshes logs', async () => {
    const triggerSync = vi.fn().mockResolvedValue({
      successMembers: 1,
      successDepartments: 0,
      failedMembers: 0,
      failedDepartments: 0,
      failures: [],
    })
    const getLogs = vi
      .fn()
      .mockResolvedValueOnce({ items: [], total: 0, page: 1, pageSize: 10 })
      .mockResolvedValueOnce({
        items: [
          { id: 'log-1', type: 'manual', result: 'success', createdAt: '2026-06-10T10:00:00Z' },
        ],
        total: 1,
        page: 1,
        pageSize: 10,
      })
    const apis = createMockApis({
      dataSourceApi: {
        getStatus: vi.fn().mockResolvedValue({ connected: true, platform: 'feishu' }),
      },
      syncApi: {
        getConfig: vi
          .fn()
          .mockResolvedValue({ enabled: true, frequencyHours: 24, startTime: '02:00' }),
        triggerSync,
        getLogs,
      },
    })
    const { result } = renderHookWithProviders(() => useDataSourcePage(apis), { apis })

    await waitForLoaded(result, 'loading')

    await act(async () => {
      await result.current.handleTriggerSync()
    })

    expect(triggerSync).toHaveBeenCalled()
    expect(toast.success).toHaveBeenCalledWith('同步完成')
  })
})

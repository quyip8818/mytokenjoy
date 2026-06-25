import { act, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useAuditSettings } from './use-audit-settings'
import { createMockApis, renderHookWithProviders } from '@/test-utils'

describe('useAuditSettings', () => {
  it('loads audit settings on mount', async () => {
    const settings = { contentRetentionEnabled: false }
    const apis = createMockApis({
      auditApi: {
        ...createMockApis().auditApi,
        getSettings: vi.fn().mockResolvedValue(settings),
        updateSettings: vi.fn().mockResolvedValue(settings),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditSettings(apis), { apis })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(apis.auditApi.getSettings).toHaveBeenCalled()
    expect(result.current.contentRetentionEnabled).toBe(false)
    expect(result.current.settings).toEqual(settings)
  })

  it('updates content retention via API', async () => {
    const apis = createMockApis({
      auditApi: {
        ...createMockApis().auditApi,
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
        updateSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: false }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditSettings(apis), { apis })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await act(async () => {
      await result.current.updateContentRetention(false)
    })

    expect(apis.auditApi.updateSettings).toHaveBeenCalledWith({ contentRetentionEnabled: false })
    expect(apis.auditApi.getSettings).toHaveBeenCalledTimes(2)
  })
})

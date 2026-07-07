import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useAuditSettings } from '@/features/audit/hooks/use-audit-settings'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useAuditSettings', () => {
  it('loads audit settings on mount', async () => {
    const settings = { contentRetentionEnabled: false }
    const apis = createMockApis({
      auditApi: {
        getSettings: vi.fn().mockResolvedValue(settings),
        updateSettings: vi.fn().mockResolvedValue(settings),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditSettings(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.auditApi.getSettings).toHaveBeenCalled()
    expect(result.current.contentRetentionEnabled).toBe(false)
    expect(result.current.settings).toEqual(settings)
  })

  it('updates content retention via API', async () => {
    const apis = createMockApis({
      auditApi: {
        getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
        updateSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: false }),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditSettings(apis), { apis })

    await waitForLoaded(result, 'loading')

    await act(async () => {
      await result.current.updateContentRetention(false)
    })

    expect(apis.auditApi.updateSettings).toHaveBeenCalledWith({ contentRetentionEnabled: false })
    expect(apis.auditApi.getSettings).toHaveBeenCalledTimes(2)
  })
})

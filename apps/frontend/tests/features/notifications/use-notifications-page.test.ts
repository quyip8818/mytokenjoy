import { describe, expect, it, vi } from 'vitest'
import { renderHookWithProviders, createMockApis } from '@tests/utils'
import { useNotificationsPage } from '@/features/notifications'

describe('useNotificationsPage', () => {
  it('returns categories and channels', () => {
    const apis = createMockApis({
      notificationApi: {
        getPreferences: vi.fn().mockResolvedValue({ preferences: [] }),
        getCapabilities: vi.fn().mockResolvedValue({
          channels: ['in_app'],
          emailConfigured: false,
          smsConfigured: false,
          inAppConfigured: true,
        }),
      },
    })

    const { result } = renderHookWithProviders(() => useNotificationsPage(apis), { apis })

    expect(result.current.categories.length).toBeGreaterThan(0)
    expect(result.current.channels.length).toBe(3)
    expect(result.current.channels.map((c) => c.key)).toEqual(['email', 'sms', 'in_app'])
  })

  it('isChannelConfigured returns false for unconfigured channels', () => {
    const apis = createMockApis({
      notificationApi: {
        getPreferences: vi.fn().mockResolvedValue({ preferences: [] }),
        getCapabilities: vi.fn().mockResolvedValue({
          channels: ['in_app'],
          emailConfigured: false,
          smsConfigured: false,
          inAppConfigured: true,
        }),
      },
    })

    const { result } = renderHookWithProviders(() => useNotificationsPage(apis), { apis })

    // in_app is always configured as a fallback
    expect(result.current.isChannelConfigured('in_app')).toBe(true)
    // email/sms not configured → initially defaults to in_app only
    expect(result.current.isChannelConfigured('email')).toBe(false)
  })

  it('isChannelEnabled uses defaults when no preferences exist', () => {
    const apis = createMockApis({
      notificationApi: {
        getPreferences: vi.fn().mockResolvedValue({ preferences: [] }),
        getCapabilities: vi.fn().mockResolvedValue({
          channels: ['in_app', 'email'],
          emailConfigured: true,
          smsConfigured: false,
          inAppConfigured: true,
        }),
      },
    })

    const { result } = renderHookWithProviders(() => useNotificationsPage(apis), { apis })

    // in_app should default to enabled for everything
    expect(result.current.isChannelEnabled('budget_alert', 'in_app')).toBe(true)
    // email defaults enabled for most categories except system_maintenance
    expect(result.current.isChannelEnabled('budget_alert', 'email')).toBe(true)
    expect(result.current.isChannelEnabled('system_maintenance', 'email')).toBe(false)
    // sms only defaults for security_event
    expect(result.current.isChannelEnabled('security_event', 'sms')).toBe(true)
    expect(result.current.isChannelEnabled('budget_alert', 'sms')).toBe(false)
  })
})

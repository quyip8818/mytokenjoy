import { useCallback, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import { useNotificationCapabilities } from './use-notification-capabilities'
import type { AppApis } from '@/api/app-apis'
import type { NotificationPreferenceEntry } from '@/api/types'

const CATEGORIES = [
  { key: 'budget_alert', label: '预算告警' },
  { key: 'key_expiration', label: 'Key 到期' },
  { key: 'usage_report', label: '用量报告' },
  { key: 'security_event', label: '安全事件' },
  { key: 'system_maintenance', label: '系统维护' },
  { key: 'overrun', label: '超支通知' },
] as const

const CHANNELS = [
  { key: 'email', label: 'Email' },
  { key: 'sms', label: 'SMS' },
  { key: 'in_app', label: '站内信' },
] as const

export function useNotificationsPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()
  const [saving, setSaving] = useState(false)

  const { data: capabilities } = useNotificationCapabilities(injectedApis)

  const {
    data: preferencesData,
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: ['notifications', 'preferences'],
    queryFn: (a) => a.notificationApi.getPreferences(),
  })

  const preferences = useMemo(() => preferencesData?.preferences ?? [], [preferencesData])

  const isChannelEnabled = useCallback(
    (category: string, channel: string): boolean => {
      const entry = preferences.find((p) => p.category === category && p.channel === channel)
      if (entry !== undefined) return entry.enabled
      // Default: true for in_app always, email for most, sms only for security
      if (channel === 'in_app') return true
      if (channel === 'email') return category !== 'system_maintenance'
      if (channel === 'sms') return category === 'security_event'
      return false
    },
    [preferences],
  )

  const isChannelConfigured = useCallback(
    (channel: string): boolean => {
      if (!capabilities) return channel === 'in_app'
      return capabilities.channels.includes(channel)
    },
    [capabilities],
  )

  const togglePreference = useCallback(
    async (category: string, channel: string, enabled: boolean) => {
      setSaving(true)
      try {
        const entry: NotificationPreferenceEntry = { category, channel, enabled }
        await apis.notificationApi.updatePreferences({ preferences: [entry] })
        queryClient.invalidateQueries({ queryKey: ['notifications', 'preferences'] })
      } catch {
        toast.error('保存偏好失败')
      } finally {
        setSaving(false)
      }
    },
    [apis, queryClient],
  )

  const resetPreferences = useCallback(async () => {
    setSaving(true)
    try {
      await apis.notificationApi.resetPreferences()
      queryClient.invalidateQueries({ queryKey: ['notifications', 'preferences'] })
      toast.success('已恢复默认偏好')
    } catch {
      toast.error('重置失败')
    } finally {
      setSaving(false)
    }
  }, [apis, queryClient])

  return {
    categories: CATEGORIES,
    channels: CHANNELS,
    loading,
    saving,
    error,
    refresh,
    isChannelEnabled,
    isChannelConfigured,
    togglePreference,
    resetPreferences,
  }
}

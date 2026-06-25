import { useCallback } from 'react'
import { auditApi } from '@/api/audit'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useAuditSettings() {
  const { data, loading, refresh } = useAsyncResource(() => auditApi.getSettings(), [])

  const updateContentRetention = useCallback(
    async (enabled: boolean) => {
      await auditApi.updateSettings({ contentRetentionEnabled: enabled })
      refresh()
    },
    [refresh],
  )

  return {
    settings: data,
    loading,
    contentRetentionEnabled: data?.contentRetentionEnabled ?? true,
    updateContentRetention,
    refresh,
  }
}

import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useAuditSettings(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const { data, loading, error, refresh } = useAsyncResource(
    () => apis.auditApi.getSettings(),
    [apis],
  )

  const updateContentRetention = useCallback(
    async (enabled: boolean) => {
      await apis.auditApi.updateSettings({ contentRetentionEnabled: enabled })
      refresh()
    },
    [apis, refresh],
  )

  return {
    settings: data,
    loading,
    error,
    contentRetentionEnabled: data?.contentRetentionEnabled ?? true,
    updateContentRetention,
    refresh,
  }
}

import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'

export function useAuditSettings(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.audit.settings(),
    queryFn: (a) => a.auditApi.getSettings(),
  })

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

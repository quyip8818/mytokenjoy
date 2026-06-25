import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { ImportResult } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useDemoCta } from '@/features/demo'
import { ROUTES } from '@/config/routes'

export function useDataSourcePage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const navigate = useNavigate()
  const credentialCta = useDemoCta('CREDENTIAL')
  const importCta = useDemoCta('IMPORT')
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [triggeringSync, setTriggeringSync] = useState(false)

  const { data, loading, error, refresh } = useAsyncResource(async () => {
    const [status, syncConfig] = await Promise.all([
      apis.dataSourceApi.getStatus(),
      apis.syncApi.getConfig(),
    ])
    return { status, syncConfig }
  }, [apis])

  const {
    data: syncLogs = [],
    loading: syncLogsLoading,
    refresh: refreshSyncLogs,
  } = useAsyncResource(async () => {
    const res = await apis.syncApi.getLogs(1, 10)
    return res.items
  }, [apis])

  const status = data?.status ?? null
  const syncConfig = data?.syncConfig ?? null
  const { openWithRefresh, open } = useWorkflowRefresh(refresh)

  const displayImportResult = importResult ?? status?.lastImportResult ?? null
  const imported = Boolean(status?.lastImport || displayImportResult)

  const handleImport = async () => {
    setImporting(true)
    try {
      const result = await apis.dataSourceApi.import()
      setImportResult(result)
      toast.success(`导入完成：${result.successMembers} 人 / ${result.successDepartments} 个部门`)
    } finally {
      setImporting(false)
    }
  }

  const handleTriggerSync = async () => {
    setTriggeringSync(true)
    try {
      const result = await apis.syncApi.triggerSync()
      setImportResult(result)
      toast.success('同步完成')
      void refreshSyncLogs()
    } finally {
      setTriggeringSync(false)
    }
  }

  const handleRetryImport = useCallback(
    async (ids: string[]) => {
      return apis.dataSourceApi.retryImport(ids)
    },
    [apis],
  )

  const openCredential = () => {
    openWithRefresh('credential-form', {
      connected: status?.connected ?? false,
      currentPlatform: status?.platform ?? null,
    })
  }

  const openSyncConfig = () => {
    open('sync-config', {
      onTriggerSync: handleTriggerSync,
      triggeringSync,
      onSuccess: refresh,
    })
  }

  const navigateToStructure = () => {
    navigate(ROUTES.orgStructure)
  }

  return {
    credentialCta,
    importCta,
    importing,
    displayImportResult,
    triggeringSync,
    status,
    syncConfig,
    loading,
    error,
    refresh,
    imported,
    setImportResult,
    handleImport,
    openCredential,
    openSyncConfig,
    navigateToStructure,
    syncLogs,
    syncLogsLoading,
    handleRetryImport,
  }
}

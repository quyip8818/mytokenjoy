import { useState } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import type { ImportResult } from '@/api/types'
import { dataSourceApi, syncApi } from '@/api/org'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useDemoCta } from '@/features/demo'
import { ROUTES } from '@/config/routes'

export function useDataSourcePage() {
  const navigate = useNavigate()
  const credentialCta = useDemoCta('CREDENTIAL')
  const importCta = useDemoCta('IMPORT')
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [triggeringSync, setTriggeringSync] = useState(false)

  const { data, loading, refresh } = useAsyncResource(async () => {
    const [status, syncConfig] = await Promise.all([dataSourceApi.getStatus(), syncApi.getConfig()])
    return { status, syncConfig }
  }, [])

  const status = data?.status ?? null
  const syncConfig = data?.syncConfig ?? null
  const { openWithRefresh, open } = useWorkflowRefresh(refresh)

  const displayImportResult = importResult ?? status?.lastImportResult ?? null
  const imported = Boolean(status?.lastImport || displayImportResult)

  const handleImport = async () => {
    setImporting(true)
    try {
      const result = await dataSourceApi.import()
      setImportResult(result)
      toast.success(`导入完成：${result.successMembers} 人 / ${result.successDepartments} 个部门`)
    } finally {
      setImporting(false)
    }
  }

  const handleTriggerSync = async () => {
    setTriggeringSync(true)
    try {
      const result = await syncApi.triggerSync()
      setImportResult(result)
      toast.success('同步完成')
    } finally {
      setTriggeringSync(false)
    }
  }

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
    imported,
    setImportResult,
    handleImport,
    openCredential,
    openSyncConfig,
    navigateToStructure,
  }
}

import { useState } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { Plug } from 'lucide-react'
import type { ImportResult, Platform } from '@/api/types'
import { dataSourceApi, syncApi } from '@/api/org'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useDemoCta } from '@/features/demo'
import { ImportResultView } from '@/components/org/import-result'
import { SyncLogTable } from '@/components/org/sync-log-table'
import { DataSourceInitProgress } from '@/components/org/data-source-init-progress'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { EmptyState } from '@/components/ui/empty-state'
import { StatusBadge } from '@/components/ui/status-badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'

const platformLabels: Record<Platform, string> = {
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企业微信',
}

export default function DataSourcePage() {
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

  if (loading || !status) {
    return (
      <PageShell>
        <DataSection loading loadingVariant="spinner">
          <div />
        </DataSection>
      </PageShell>
    )
  }

  const imported = Boolean(status.lastImport || displayImportResult)

  return (
    <PageShell>
      <DataSourceInitProgress connected={status.connected} imported={imported} />

      {status.connected && status.platform ? (
        <div className="flex items-center justify-between gap-4 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3">
          <div className="flex items-center gap-2">
            <StatusBadge variant="success">已连接</StatusBadge>
            <span className="text-sm text-emerald-800">
              当前数据源：{platformLabels[status.platform]}
            </span>
          </div>
          <div className="flex gap-2">
            <PermissionGate write>
              <Button
                id={credentialCta.id}
                variant="outline"
                size="sm"
                className={cn(credentialCta.className)}
                onClick={openCredential}
              >
                配置凭证
              </Button>
              <Button variant="outline" size="sm" onClick={openSyncConfig}>
                编辑同步策略
              </Button>
            </PermissionGate>
          </div>
        </div>
      ) : (
        <PermissionGate write permission={PERMISSION.ORG_DATASOURCE}>
          <EmptyState
            icon={Plug}
            title="尚未连接第三方平台"
            description="连接后可导入组织"
            actionLabel="配置凭证"
            onAction={openCredential}
            actionClassName={credentialCta.className}
            actionId={credentialCta.id}
          />
        </PermissionGate>
      )}

      {status.connected && (
        <>
          <DataSection
            title="全量导入"
            headerAction={
              <PermissionGate write permission={PERMISSION.ORG_DATASOURCE}>
                <Button
                  id={importCta.id}
                  size="sm"
                  variant="brand"
                  className={cn(importCta.className)}
                  onClick={handleImport}
                  disabled={importing}
                >
                  {importing ? '导入中...' : '执行全量导入'}
                </Button>
              </PermissionGate>
            }
          >
            {displayImportResult ? (
              <ImportResultView
                result={displayImportResult}
                onUpdate={setImportResult}
                onNavigateOrg={() => navigate('/org/structure')}
              />
            ) : (
              <p className="text-sm text-muted-foreground">
                从已连接平台拉取组织与成员数据，完成后可在组织架构中查看。
              </p>
            )}
          </DataSection>

          <DataSection
            title="同步策略"
            headerAction={
              <PermissionGate write permission={PERMISSION.ORG_DATASOURCE}>
                <Button variant="outline" size="sm" onClick={openSyncConfig}>
                  编辑
                </Button>
              </PermissionGate>
            }
            loading={!syncConfig}
            loadingVariant="spinner"
          >
            {syncConfig && (
              <div className="space-y-1 text-sm text-muted-foreground">
                <p>
                  {syncConfig.enabled ? '已启用' : '未启用'} · 每 {syncConfig.frequencyHours}h ·{' '}
                  {syncConfig.startTime} 起
                </p>
                <p>
                  删除保护：成员 {syncConfig.deleteMemberThreshold} 人 / 部门{' '}
                  {syncConfig.deleteDepartmentThreshold} 个
                </p>
              </div>
            )}
          </DataSection>

          <DataSection title="同步记录">
            <SyncLogTable />
          </DataSection>
        </>
      )}
    </PageShell>
  )
}

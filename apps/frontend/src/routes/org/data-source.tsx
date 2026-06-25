import { Plug } from 'lucide-react'
import { ImportResultView } from '@/components/org/import-result'
import { SyncLogTable } from '@/components/org/sync-log-table'
import { DataSourceInitProgress } from '@/components/org/data-source-init-progress'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
import { EmptyState } from '@/components/ui/empty-state'
import { StatusBadge } from '@/components/ui/status-badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { PLATFORM_LABELS } from '@/lib/labels'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { useDataSourcePage } from '@/routes/org/hooks/use-data-source-page'

export default function DataSourcePage() {
  const {
    credentialCta,
    importCta,
    importing,
    displayImportResult,
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
  } = useDataSourcePage()

  if (loading) {
    return (
      <PageShell>
        <DataSection loading loadingVariant="spinner">
          <div />
        </DataSection>
      </PageShell>
    )
  }

  if (error) {
    return (
      <PageShell>
        <ErrorState message={error.message} onRetry={refresh} />
      </PageShell>
    )
  }

  if (!status) {
    return (
      <PageShell>
        <DataSection loading loadingVariant="spinner">
          <div />
        </DataSection>
      </PageShell>
    )
  }

  return (
    <PageShell>
      <DataSourceInitProgress connected={status.connected} imported={imported} />

      {status.connected && status.platform ? (
        <div className="flex items-center justify-between gap-4 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3">
          <div className="flex items-center gap-2">
            <StatusBadge variant="success">已连接</StatusBadge>
            <span className="text-sm text-emerald-800">
              当前数据源：{PLATFORM_LABELS[status.platform]}
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
                onNavigateOrg={navigateToStructure}
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

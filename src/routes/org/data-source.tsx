import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { Plug } from 'lucide-react'
import type { DataSourceStatus, ImportResult, Platform, SyncConfig } from '@/api/types'
import { dataSourceApi, syncApi } from '@/api/org'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { useDemoCta } from '@/features/demo'
import { ImportResultView } from '@/components/org/import-result'
import { SyncLogTable } from '@/components/org/sync-log-table'
import { DataSourceInitProgress } from '@/components/org/data-source-init-progress'
import { EmptyState } from '@/components/ui/empty-state'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const platformLabels: Record<Platform, string> = {
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企业微信',
}

export default function DataSourcePage() {
  const navigate = useNavigate()
  const { open } = useWorkflow()
  const credentialCta = useDemoCta('CREDENTIAL')
  const importCta = useDemoCta('IMPORT')
  const [status, setStatus] = useState<DataSourceStatus | null>(null)
  const [syncConfig, setSyncConfig] = useState<SyncConfig | null>(null)
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [triggeringSync, setTriggeringSync] = useState(false)

  const loadStatus = useCallback(async () => {
    const [s, cfg] = await Promise.all([dataSourceApi.getStatus(), syncApi.getConfig()])
    setStatus(s)
    setSyncConfig(cfg)
    if (s.lastImportResult) setImportResult(s.lastImportResult)
  }, [])

  useEffect(() => {
    void Promise.all([dataSourceApi.getStatus(), syncApi.getConfig()]).then(([s, cfg]) => {
      setStatus(s)
      setSyncConfig(cfg)
      if (s.lastImportResult) setImportResult(s.lastImportResult)
    })
  }, [])

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
    open('credential-form', {
      connected: status?.connected ?? false,
      currentPlatform: status?.platform ?? null,
      onSuccess: loadStatus,
    })
  }

  const openSyncConfig = () => {
    open('sync-config', {
      onTriggerSync: handleTriggerSync,
      triggeringSync,
      onSuccess: loadStatus,
    })
  }

  if (!status) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">加载中...</p>
      </div>
    )
  }

  const imported = Boolean(status.lastImport || importResult)

  return (
    <div className="space-y-6">
      <DataSourceInitProgress connected={status.connected} imported={imported} />

      {status.connected && status.platform ? (
        <div className="flex items-center justify-between gap-4 px-4 py-3 bg-green-50 border border-green-200 rounded-md">
          <div className="flex items-center gap-2">
            <Badge variant="default" className="bg-green-600">
              已连接
            </Badge>
            <span className="text-sm text-green-800">
              当前数据源：{platformLabels[status.platform]}
            </span>
          </div>
          <div className="flex gap-2">
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
          </div>
        </div>
      ) : (
        <EmptyState
          icon={Plug}
          title="尚未连接第三方平台"
          description="连接后可导入组织"
          actionLabel="配置凭证"
          onAction={openCredential}
          actionClassName={credentialCta.className}
          actionId={credentialCta.id}
        />
      )}

      {status.connected && (
        <>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle>全量导入</CardTitle>
              <Button
                id={importCta.id}
                className={cn(importCta.className)}
                onClick={handleImport}
                disabled={importing}
              >
                {importing ? '导入中...' : '执行全量导入'}
              </Button>
            </CardHeader>
            <CardContent>
              {importResult && (
                <ImportResultView
                  result={importResult}
                  onUpdate={setImportResult}
                  onNavigateOrg={() => navigate('/org/structure')}
                />
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle>同步策略</CardTitle>
              <Button variant="outline" size="sm" onClick={openSyncConfig}>
                编辑
              </Button>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground space-y-1">
              {syncConfig ? (
                <>
                  <p>
                    {syncConfig.enabled ? '已启用' : '未启用'} · 每 {syncConfig.frequencyHours}h ·{' '}
                    {syncConfig.startTime} 起
                  </p>
                  <p>
                    删除保护：成员 {syncConfig.deleteMemberThreshold} 人 / 部门{' '}
                    {syncConfig.deleteDepartmentThreshold} 个
                  </p>
                </>
              ) : (
                <p>加载中...</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>同步记录</CardTitle>
            </CardHeader>
            <CardContent>
              <SyncLogTable />
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}

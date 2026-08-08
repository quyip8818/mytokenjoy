import { Power, Upload, Pencil, Plus, RefreshCw, Layers } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ActionIcon } from '@/components/ui/action-icon'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { ModelTable } from '@/features/models'
import type { usePlatformModelsPage } from '../hooks/use-platform-models-page'

type Props = ReturnType<typeof usePlatformModelsPage>

export function PlatformModelsPageShell({
  models,
  loading,
  error,
  refresh,
  publishing,
  syncing,
  handlePublish,
  handleSync,
  handleToggle,
  openCreate,
  openEdit,
  openChannels,
}: Props) {
  return (
    <PageShell>
      <PageHeader
        testId="page-platform-models"
        title="模型目录"
        description={`共 ${models.length} 个全局模型`}
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" disabled={syncing} onClick={handleSync}>
              <RefreshCw className={`mr-1.5 h-4 w-4 ${syncing ? 'animate-spin' : ''}`} />
              {syncing ? '同步中...' : '同步模型'}
            </Button>
            <Button onClick={openCreate}>
              <Plus className="mr-1.5 h-4 w-4" />
              添加模型
            </Button>
            <Button variant="brand" disabled={publishing} onClick={handlePublish}>
              <Upload className="mr-1.5 h-4 w-4" />
              {publishing ? '发布中...' : '发布'}
            </Button>
          </div>
        }
      />

      <Card className="border-border shadow-xs">
        <CardContent className="px-5 pt-5 pb-4">
          <DataSection loading={loading} error={error} onRetry={refresh} skeletonColumns={6}>
            <ModelTable
              models={models}
              extraColumns={[
                {
                  header: '状态',
                  render: (m) => (
                    <Badge variant={m.deprecated ? 'outline' : 'default'}>
                      {m.deprecated ? '已下线' : '启用'}
                    </Badge>
                  ),
                },
              ]}
              renderActions={(m) => (
                <div className="inline-flex items-center gap-1.5">
                  <ActionIcon hint="查看渠道" onClick={() => openChannels(m)}>
                    <Layers className="h-5 w-5" />
                  </ActionIcon>
                  <ActionIcon hint="编辑模型" onClick={() => openEdit(m)}>
                    <Pencil className="h-5 w-5" />
                  </ActionIcon>
                  <ActionIcon
                    hint={m.deprecated ? '恢复' : '下线'}
                    onClick={() => handleToggle(m)}
                    className={m.deprecated ? 'text-green-500' : 'text-amber-500'}
                  >
                    <Power className="h-5 w-5" />
                  </ActionIcon>
                </div>
              )}
            />
          </DataSection>
        </CardContent>
      </Card>
    </PageShell>
  )
}

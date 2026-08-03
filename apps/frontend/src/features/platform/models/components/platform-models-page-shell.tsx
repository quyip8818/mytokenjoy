import { Power, Upload, Pencil, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
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
  handlePublish,
  handleToggle,
  openCreate,
  openEdit,
}: Props) {
  return (
    <PageShell>
      <PageHeader
        title="模型目录"
        description={`共 ${models.length} 个全局模型`}
        actions={
          <div className="flex items-center gap-2">
            <Button size="sm" onClick={openCreate}>
              <Plus className="mr-1.5 h-4 w-4" />
              添加模型
            </Button>
            <Button size="sm" variant="brand" disabled={publishing} onClick={handlePublish}>
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
                <div className="inline-flex items-center gap-1">
                  <button
                    onClick={() => openEdit(m)}
                    className="rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
                    title="编辑模型"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </button>
                  <button
                    onClick={() => handleToggle(m)}
                    className={`rounded p-1.5 hover:bg-muted ${m.deprecated ? 'text-green-500' : 'text-amber-500'}`}
                    title={m.deprecated ? '恢复' : '下线'}
                  >
                    <Power className="h-3.5 w-3.5" />
                  </button>
                </div>
              )}
            />
          </DataSection>
        </CardContent>
      </Card>
    </PageShell>
  )
}

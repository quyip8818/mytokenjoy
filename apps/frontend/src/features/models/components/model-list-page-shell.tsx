import { Box } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import type { useModelListPage } from '@/features/models'
import { ModelListTable } from './model-list-table'

type ModelListPageShellProps = ReturnType<typeof useModelListPage>

export function ModelListPageShell({
  models,
  counts,
  tab,
  setTab,
  loading,
  error,
  refresh,
  canManage,
  modelCta,
  rowClass,
  handleToggle,
  handleDelete,
  openCreate,
  openEdit,
}: ModelListPageShellProps) {
  return (
    <PageShell
      actions={
        <PermissionGate write permission={PERMISSION.MODEL_MANAGE}>
          <Button
            id={modelCta.id}
            size="sm"
            variant="brand"
            className={modelCta.className}
            onClick={openCreate}
          >
            添加模型
          </Button>
        </PermissionGate>
      }
    >
      <Tabs value={tab} onValueChange={(value) => setTab(value as typeof tab)}>
        <Card className="border-border shadow-xs">
          <CardContent className="px-5 pt-4 pb-4">
            <TabsList variant="line" className="mb-4">
              <TabsTrigger value="all">全部模型 ({counts.all})</TabsTrigger>
              <TabsTrigger value="custom">自定义模型 ({counts.custom})</TabsTrigger>
              <TabsTrigger value="builtin">内置模型 ({counts.builtin})</TabsTrigger>
            </TabsList>

            <TabsContent value={tab} className="mt-0">
              <DataSection
                loading={loading}
                error={error}
                onRetry={refresh}
                skeletonColumns={7}
                className="border-0 shadow-none"
                contentClassName="p-0"
                empty={listEmpty(loading, models, {
                  icon: Box,
                  title: '暂无模型',
                  description: '添加自定义模型以扩展可用模型列表',
                  actionLabel: canManage ? '添加模型' : undefined,
                  onAction: canManage ? openCreate : undefined,
                })}
              >
                <ModelListTable
                  models={models}
                  canManage={canManage}
                  showActions={tab !== 'builtin'}
                  rowClass={rowClass}
                  onToggle={handleToggle}
                  onEdit={openEdit}
                  onDelete={handleDelete}
                />
              </DataSection>
            </TabsContent>
          </CardContent>
        </Card>
      </Tabs>
    </PageShell>
  )
}

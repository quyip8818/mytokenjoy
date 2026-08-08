import { Box, Cpu, Layers } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import { IS_SAAS } from '@/config/app'
import type { useModelListPage } from '@/features/models'
import { ModelListTable } from './model-list-table'

type ModelListPageShellProps = ReturnType<typeof useModelListPage>

function StatChip({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ElementType
  label: string
  value: number
}) {
  return (
    <div className="flex items-center gap-2 rounded-lg bg-muted/60 px-3 py-1.5">
      <Icon className="size-3.5 text-muted-foreground" />
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-xs font-semibold tabular-nums text-foreground">{value}</span>
    </div>
  )
}

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
  openCreate,
  openEdit,
  discountMap,
}: ModelListPageShellProps) {
  const isLocalDeploy = !IS_SAAS
  const tableContent = (
    <DataSection
      loading={loading}
      error={error}
      onRetry={refresh}
      skeletonColumns={6}
      empty={listEmpty(loading, models, {
        icon: Box,
        title: '暂无模型',
        description: isLocalDeploy ? '添加自定义模型以扩展可用模型列表' : '当前没有可用的内置模型',
        actionLabel: isLocalDeploy && canManage ? '添加模型' : undefined,
        onAction: isLocalDeploy && canManage ? openCreate : undefined,
      })}
    >
      <ModelListTable
        models={models}
        canManage={canManage}
        discountMap={discountMap}
        rowClass={rowClass}
        onToggle={handleToggle}
        onEdit={openEdit}
      />
    </DataSection>
  )

  // SaaS version: simple table without tabs
  if (!isLocalDeploy) {
    return (
      <PageShell>
        <PageHeader testId="page-models-list" title="模型列表" />
        <Card className="min-h-[360px] border-border shadow-xs">
          <CardContent className="px-5 pt-5 pb-4">{tableContent}</CardContent>
        </Card>
      </PageShell>
    )
  }

  // Self-hosted version: stats + tabs + add button
  return (
    <PageShell>
      <PageHeader
        testId="page-models-list"
        title="模型列表"
        actions={
          <PermissionGate write permission={PERMISSION.MODEL_MANAGE}>
            <Button
              id={modelCta.id}

              variant="brand"
              className={modelCta.className}
              onClick={openCreate}
            >
              添加模型
            </Button>
          </PermissionGate>
        }
      />

      {/* Stats bar */}
      {!loading && models.length > 0 && (
        <div className="flex flex-wrap items-center gap-2">
          <StatChip icon={Layers} label="总计" value={counts.all} />
          <StatChip icon={Cpu} label="自定义" value={counts.custom} />
          <StatChip icon={Box} label="内置" value={counts.builtin} />
        </div>
      )}

      <Tabs value={tab} onValueChange={(value) => setTab(value as typeof tab)}>
        <Card className="min-h-[360px] border-border shadow-xs">
          <CardContent className="px-5 pt-5 pb-4">
            <TabsList variant="line" className="mb-4">
              <TabsTrigger value="all">全部</TabsTrigger>
              <TabsTrigger value="custom">自定义</TabsTrigger>
              <TabsTrigger value="builtin">内置</TabsTrigger>
            </TabsList>

            <TabsContent value={tab} className="mt-0">
              {tableContent}
            </TabsContent>
          </CardContent>
        </Card>
      </Tabs>
    </PageShell>
  )
}

import { Box } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import type { useModelListPage } from '@/features/models'
import { ModelListTable } from './model-list-table'

type ModelListPageShellProps = ReturnType<typeof useModelListPage>

export function ModelListPageShell({
  models,
  loading,
  error,
  refresh,
  canManage,
  modelCta,
  rowClass,
  handleToggle,
  openCreate,
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
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={7}
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
          rowClass={rowClass}
          onToggle={handleToggle}
        />
      </DataSection>
    </PageShell>
  )
}

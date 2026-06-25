import { Wallet } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { BudgetGroupTable } from '@/routes/budget/components/budget-group-table'
import { useBudgetAllocationPage } from '@/routes/budget/hooks/use-budget-allocation-page'

export default function BudgetAllocationPage() {
  const { groups, loading, error, refresh, canWrite, rowClass, handleDelete, openForm } =
    useBudgetAllocationPage()

  return (
    <PageShell
      description={
        <p className="text-sm text-muted-foreground">
          预算总览管理组织树逐级分配；本页管理独立于组织树的 Budget Group（虚拟项目组）。
        </p>
      }
      actions={
        <PermissionGate write permission={PERMISSION.BUDGET_ALLOCATE}>
          <Button size="sm" variant="brand" onClick={() => openForm()}>
            新建预算组
          </Button>
        </PermissionGate>
      }
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={6}
        empty={listEmpty(loading, groups, {
          icon: Wallet,
          title: '暂无预算组',
          description: '创建预算组以管理虚拟项目组的独立预算',
          actionLabel: '新建预算组',
          onAction: () => openForm(),
        })}
      >
        <BudgetGroupTable
          groups={groups}
          canWrite={canWrite}
          rowClass={rowClass}
          onEdit={openForm}
          onDelete={handleDelete}
        />
      </DataSection>
    </PageShell>
  )
}

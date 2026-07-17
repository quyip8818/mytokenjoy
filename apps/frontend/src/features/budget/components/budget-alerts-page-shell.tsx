import { Bell, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import type { useBudgetAlertRulesPage } from '@/features/budget'
import { AlertRuleDialog } from './alert-rule-dialog'
import { BudgetAlertsTable } from './budget-alerts-table'
import { BudgetAlertsStats } from './budget-alerts-stats'
import { BudgetAlertsToolbar } from './budget-alerts-toolbar'

type BudgetAlertsPageShellProps = ReturnType<typeof useBudgetAlertRulesPage>

export function BudgetAlertsPageShell({
  rules,
  allRules,
  projects,
  tree,
  roles,
  stats,
  loading,
  error,
  refresh,
  dialogOpen,
  setDialogOpen,
  editingRule,
  deleteTarget,
  setDeleteTarget,
  handleToggle,
  handleDelete,
  openCreate,
  openEdit,
  saveRule,
  typeFilter,
  setTypeFilter,
  statusFilter,
  setStatusFilter,
  search,
  setSearch,
}: BudgetAlertsPageShellProps) {
  const hasRules = allRules.length > 0

  return (
    <PageShell
      actions={
        <Button size="sm" className="gap-1.5" onClick={openCreate}>
          <Plus className="size-3.5" />
          创建规则
        </Button>
      }
    >
      <DataSection loading={loading} error={error} onRetry={() => void refresh()}>
        {!hasRules ? (
          <EmptyState onCreate={openCreate} />
        ) : (
          <div className="space-y-5">
            <BudgetAlertsStats stats={stats} />
            <BudgetAlertsToolbar
              typeFilter={typeFilter}
              onTypeFilterChange={setTypeFilter}
              statusFilter={statusFilter}
              onStatusFilterChange={setStatusFilter}
              search={search}
              onSearchChange={setSearch}
            />
            <BudgetAlertsTable
              rules={rules}
              projects={projects}
              roles={roles}
              onToggle={(rule) => void handleToggle(rule)}
              onEdit={openEdit}
              onDelete={setDeleteTarget}
            />
          </div>
        )}
      </DataSection>
      <AlertRuleDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        rule={editingRule}
        tree={tree}
        projects={projects}
        roles={roles}
        onSave={saveRule}
      />
      <ConfirmActionDialog
        state={
          deleteTarget
            ? {
                open: true,
                title: '删除预警规则',
                desc: `确定删除「${deleteTarget.targetName}」的预警规则？此操作不可撤销。`,
                variant: 'danger',
                confirmLabel: '删除',
                onConfirm: () => void handleDelete(),
              }
            : null
        }
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        onClose={() => setDeleteTarget(null)}
      />
    </PageShell>
  )
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed border-border py-16">
      <span className="flex size-12 items-center justify-center rounded-xl bg-primary/8 text-primary">
        <Bell className="size-6" strokeWidth={1.5} />
      </span>
      <div className="text-center">
        <p className="text-sm font-medium text-foreground">尚未配置预警规则</p>
        <p className="mt-1 text-xs text-muted-foreground">
          设置预警阈值，在预算即将超支时及时通知相关负责人
        </p>
      </div>
      <Button size="sm" className="gap-1.5" onClick={onCreate}>
        <Plus className="size-3.5" />
        创建第一条规则
      </Button>
    </div>
  )
}

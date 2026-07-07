import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import type { useBudgetAlertRulesPage } from '@/features/budget/hooks/use-budget-alert-rules-page'
import { AlertRuleDialog } from './alert-rule-dialog'
import { BudgetAlertsTable } from './budget-alerts-table'

type BudgetAlertsPageShellProps = ReturnType<typeof useBudgetAlertRulesPage>

export function BudgetAlertsPageShell({
  rules,
  projects,
  tree,
  roles,
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
}: BudgetAlertsPageShellProps) {
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
        <BudgetAlertsTable
          rules={rules}
          projects={projects}
          onToggle={(rule) => void handleToggle(rule)}
          onEdit={openEdit}
          onDelete={setDeleteTarget}
        />
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

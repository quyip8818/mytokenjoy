import { Bell, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { EmptyState } from '@/components/ui/empty-state'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
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
    <PageShell>
      <PageHeader
        title="预警规则"
        description="设置预警阈值，在预算即将超支时及时通知相关负责人"
        actions={
          <PermissionGate permission={PERMISSION.BUDGET_ADMIN}>
            <Button size="sm" className="gap-1.5" onClick={openCreate}>
              <Plus className="size-3.5" />
              创建规则
            </Button>
          </PermissionGate>
        }
      />

      <DataSection loading={loading} error={error} onRetry={() => void refresh()}>
        {!hasRules ? (
          <EmptyState
            icon={Bell}
            title="尚未配置预警规则"
            description="设置预警阈值，在预算即将超支时及时通知相关负责人"
            actionLabel="创建第一条规则"
            onAction={openCreate}
          />
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
            <Card className="border-border shadow-xs">
              <CardContent className="px-5 pt-5 pb-4">
                <BudgetAlertsTable
                  rules={rules}
                  projects={projects}
                  roles={roles}
                  onToggle={(rule) => void handleToggle(rule)}
                  onEdit={openEdit}
                  onDelete={setDeleteTarget}
                />
              </CardContent>
            </Card>
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

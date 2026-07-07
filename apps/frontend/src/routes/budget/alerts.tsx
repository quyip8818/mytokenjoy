import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { cn } from '@/lib/utils'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { AlertRuleDialog, useBudgetAlertRulesPage } from '@/features/budget'

function thresholdClass(threshold: number) {
  if (threshold >= 100) return 'bg-red-50 text-red-700 border-red-200'
  if (threshold >= 90) return 'bg-amber-50 text-amber-700 border-amber-200'
  return 'bg-emerald-50 text-emerald-700 border-emerald-200'
}

export default function BudgetAlertsPage() {
  const {
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
  } = useBudgetAlertRulesPage()

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
        <div className="rounded-lg border border-border shadow-xs">
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  监控对象
                </TableHead>
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  类型
                </TableHead>
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  阈值
                </TableHead>
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  通知角色
                </TableHead>
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  状态
                </TableHead>
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rules.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-8 text-center text-sm text-muted-foreground">
                    暂无预警规则，点击上方按钮创建
                  </TableCell>
                </TableRow>
              ) : (
                rules.map((rule) => (
                  <TableRow
                    key={rule.id}
                    className="border-border-subtle even:bg-muted/40 hover:bg-muted/50"
                  >
                    <TableCell className="font-medium">
                      {rule.targetType === 'project' ? (
                        <div className="flex items-center gap-1.5">
                          <span className="text-muted-foreground">
                            {projects.find((project) => project.id === rule.targetId)
                              ?.departmentName ?? ''}
                          </span>
                          <span className="text-muted-foreground">/</span>
                          <span>{rule.targetName}</span>
                        </div>
                      ) : (
                        rule.targetName
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={cn(
                          rule.targetType === 'team'
                            ? 'border-border text-muted-foreground'
                            : 'border-primary/20 text-primary',
                        )}
                      >
                        {rule.targetType === 'team' ? '团队' : '项目'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {rule.thresholds.map((threshold) => (
                          <Badge
                            key={threshold}
                            variant="outline"
                            className={cn('tabular-nums', thresholdClass(threshold))}
                          >
                            {threshold}%
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {rule.notifyRoleIds.length} 个角色
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={rule.enabled}
                        onCheckedChange={() => void handleToggle(rule)}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-7"
                          aria-label="编辑规则"
                          onClick={() => openEdit(rule)}
                        >
                          <Pencil className="size-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-7 text-destructive"
                          aria-label="删除规则"
                          onClick={() => setDeleteTarget(rule)}
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
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

      <AlertDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除预警规则</AlertDialogTitle>
            <AlertDialogDescription>
              确定删除「{deleteTarget?.targetName}」的预警规则？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setDeleteTarget(null)}>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => void handleDelete()}
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageShell>
  )
}

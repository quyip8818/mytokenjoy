import { useState, useEffect } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { budgetApi } from '@/api/budget'
import type { AlertRule, BudgetProject } from '@/api/types'
import { AlertRuleDialog } from '@/components/budget/alert-rule-dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'
import { Pencil, Plus, Trash2 } from 'lucide-react'

function thresholdClass(t: number) {
  if (t >= 100) return 'bg-red-50 text-red-700 border-red-200'
  if (t >= 90) return 'bg-amber-50 text-amber-700 border-amber-200'
  return 'bg-emerald-50 text-emerald-700 border-emerald-200'
}

export default function BudgetAlertsPage() {
  const [rules, setRules] = useState<AlertRule[]>([])
  const [projects, setProjects] = useState<BudgetProject[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AlertRule | null>(null)

  useEffect(() => {
    budgetApi.getAlerts().then(setRules)
    budgetApi.getProjects().then(setProjects)
  }, [])

  const reload = () => {
    budgetApi.getAlerts().then(setRules)
    budgetApi.getProjects().then(setProjects)
  }

  const handleToggle = async (rule: AlertRule) => {
    await budgetApi.updateAlert(rule.id, { enabled: !rule.enabled })
    setRules(rules.map(r => r.id === rule.id ? { ...r, enabled: !r.enabled } : r))
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    await budgetApi.deleteAlert(deleteTarget.id)
    setDeleteTarget(null)
    reload()
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div />
        <Button size="sm" className="gap-1.5" onClick={() => { setEditingRule(null); setDialogOpen(true) }}>
          <Plus className="size-3.5" />
          创建规则
        </Button>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border shadow-xs">
        <Table>
          <TableHeader>
            <TableRow className="border-border/50 hover:bg-transparent">
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">监控对象</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">类型</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">阈值</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">通知角色</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">状态</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">操作</TableHead>
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
                <TableRow key={rule.id} className="border-border-subtle even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="font-medium">
                    {rule.targetType === 'project' ? (
                      <div className="flex items-center gap-1.5">
                        <span className="text-muted-foreground">{projects.find(p => p.id === rule.targetId)?.departmentName ?? ''}</span>
                        <span className="text-muted-foreground">/</span>
                        <span>{rule.targetName}</span>
                      </div>
                    ) : (
                      rule.targetName
                    )}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={cn(
                      rule.targetType === 'team'
                        ? 'border-border text-muted-foreground'
                        : 'border-primary/20 text-primary'
                    )}>
                      {rule.targetType === 'team' ? '团队' : '项目'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      {rule.thresholds.map((t) => (
                        <Badge key={t} variant="outline" className={cn('tabular-nums', thresholdClass(t))}>{t}%</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {rule.notifyRoleIds.length} 个角色
                  </TableCell>
                  <TableCell>
                    <Switch checked={rule.enabled} onCheckedChange={() => handleToggle(rule)} />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7"
                        aria-label="编辑规则"
                        onClick={() => { setEditingRule(rule); setDialogOpen(true) }}
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

      {/* Create/Edit Dialog */}
      <AlertRuleDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        rule={editingRule}
        onSaved={reload}
      />

      {/* Delete confirmation */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除预警规则</AlertDialogTitle>
            <AlertDialogDescription>
              确定删除「{deleteTarget?.targetName}」的预警规则？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setDeleteTarget(null)}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-white hover:bg-destructive/90">删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

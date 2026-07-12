import type { ProjectView } from '@/api/types'
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
import { cn } from '@/lib/utils'
import { Pencil, Trash2 } from 'lucide-react'
import { thresholdClass, type AlertRuleView } from '../lib/alerts'

interface BudgetAlertsTableProps {
  rules: AlertRuleView[]
  projects: ProjectView[]
  onToggle: (rule: AlertRuleView) => void
  onEdit: (rule: AlertRuleView) => void
  onDelete: (rule: AlertRuleView) => void
}

export function BudgetAlertsTable({
  rules,
  projects,
  onToggle,
  onEdit,
  onDelete,
}: BudgetAlertsTableProps) {
  return (
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
                        {projects.find((project) => project.id === rule.targetId)?.departmentName ??
                          ''}
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
                  <Switch checked={rule.enabled} onCheckedChange={() => onToggle(rule)} />
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      aria-label="编辑规则"
                      onClick={() => onEdit(rule)}
                    >
                      <Pencil className="size-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7 text-destructive"
                      aria-label="删除规则"
                      onClick={() => onDelete(rule)}
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
  )
}

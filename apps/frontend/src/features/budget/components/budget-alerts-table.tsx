import type { ProjectView, Role } from '@/api/types'
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { MoreHorizontal, Pencil, Trash2, Ban, CheckCircle2, ArrowDownUp } from 'lucide-react'
import { thresholdClass, type AlertRuleView } from '../lib/alerts'
import { POLICY_LABELS } from '../lib/constants'

interface BudgetAlertsTableProps {
  rules: AlertRuleView[]
  projects: ProjectView[]
  roles: Role[]
  onToggle: (rule: AlertRuleView) => void
  onEdit: (rule: AlertRuleView) => void
  onDelete: (rule: AlertRuleView) => void
}

const ACTION_ICONS: Record<string, React.ElementType> = {
  hard_reject: Ban,
  approval: CheckCircle2,
  downgrade: ArrowDownUp,
}

export function BudgetAlertsTable({
  rules,
  projects,
  roles,
  onToggle,
  onEdit,
  onDelete,
}: BudgetAlertsTableProps) {
  const roleMap = new Map(roles.map((r) => [r.id, r.name]))

  return (
    <div className="rounded-xl border border-border shadow-xs">
      <Table>
        <TableHeader>
          <TableRow className="border-border/50 hover:bg-transparent">
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              监控对象
            </TableHead>
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              阈值
            </TableHead>
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              触发动作
            </TableHead>
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              通知角色
            </TableHead>
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              状态
            </TableHead>
            <TableHead className="w-12" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rules.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="py-12 text-center text-sm text-muted-foreground">
                没有匹配的预警规则
              </TableCell>
            </TableRow>
          ) : (
            rules.map((rule) => {
              const ActionIcon = ACTION_ICONS[rule.action] ?? Ban
              const roleNames = rule.notifyRoleIds
                .map((id) => roleMap.get(id))
                .filter(Boolean) as string[]

              return (
                <TableRow
                  key={rule.id}
                  className={cn(
                    'border-border/50 hover:bg-muted/50',
                    !rule.enabled && 'opacity-50',
                  )}
                >
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-foreground">
                        {rule.targetType === 'project' ? (
                          <>
                            <span className="text-muted-foreground">
                              {projects.find((p) => p.id === rule.targetId)?.departmentName ?? ''}
                            </span>
                            <span className="mx-1 text-muted-foreground">/</span>
                            {rule.targetName}
                          </>
                        ) : (
                          rule.targetName
                        )}
                      </span>
                      <Badge
                        variant="outline"
                        className={cn(
                          'text-[10px]',
                          rule.targetType === 'team'
                            ? 'border-border text-muted-foreground'
                            : 'border-primary/20 text-primary',
                        )}
                      >
                        {rule.targetType === 'team' ? '团队' : '项目'}
                      </Badge>
                    </div>
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
                  <TableCell>
                    <div className="flex items-center gap-1.5">
                      <ActionIcon className="size-3.5" />
                      <span className="text-sm">
                        {POLICY_LABELS[rule.action]?.label ?? '硬拒绝'}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <RoleDisplay roleNames={roleNames} />
                  </TableCell>
                  <TableCell>
                    <Switch checked={rule.enabled} onCheckedChange={() => onToggle(rule)} />
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-7"
                          aria-label="更多操作"
                        >
                          <MoreHorizontal className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => onEdit(rule)}>
                          <Pencil className="mr-2 size-3.5" />
                          编辑
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => onDelete(rule)}
                          className="text-destructive focus:text-destructive"
                        >
                          <Trash2 className="mr-2 size-3.5" />
                          删除
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              )
            })
          )}
        </TableBody>
      </Table>
    </div>
  )
}

function RoleDisplay({ roleNames }: { roleNames: string[] }) {
  if (roleNames.length === 0) {
    return <span className="text-sm text-muted-foreground">—</span>
  }
  const visible = roleNames.slice(0, 2)
  const rest = roleNames.length - 2

  return (
    <div className="flex items-center gap-1">
      {visible.map((name) => (
        <Badge key={name} variant="secondary" className="text-xs font-normal">
          {name}
        </Badge>
      ))}
      {rest > 0 && (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Badge variant="secondary" className="cursor-default text-xs font-normal">
                +{rest}
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">{roleNames.slice(2).join('、')}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    </div>
  )
}

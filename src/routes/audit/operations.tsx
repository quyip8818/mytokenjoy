import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { auditApi } from '@/api/audit'
import type { OperationLog } from '@/api/types'

const actionLabels: Record<string, string> = {
  key_create: 'Key 创建',
  key_disable: 'Key 禁用',
  key_rotate: 'Key 轮转',
  budget_change: '预算变更',
  budget_approve: '预算审批',
  permission_change: '权限变更',
  role_assign: '角色分配',
  model_whitelist_change: '白名单变更',
  member_add: '成员添加',
  member_remove: '成员移除',
  org_structure_change: '组织结构变更',
}

function getActionStyle(action: string): string {
  if (action.startsWith('key_')) return 'bg-amber-50 text-amber-700'
  if (action.startsWith('budget_')) return 'bg-emerald-50 text-emerald-700'
  if (action.startsWith('permission') || action.startsWith('role'))
    return 'bg-indigo-50 text-indigo-700'
  return 'bg-slate-100 text-slate-600'
}

export default function OperationLogsPage() {
  const [logs, setLogs] = useState<OperationLog[]>([])
  const [actionFilter, setActionFilter] = useState<string>('all')

  useEffect(() => {
    const params = actionFilter !== 'all' ? { action: actionFilter } : undefined
    auditApi.getOperations(params).then((res) => setLogs(res.items))
  }, [actionFilter])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end">
        <Select value={actionFilter} onValueChange={(v) => setActionFilter(v ?? 'all')}>
          <SelectTrigger className="w-40 border-border/60 focus:ring-indigo-500">
            <SelectValue placeholder="全部类型" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部类型</SelectItem>
            <SelectItem value="key_create">Key 创建</SelectItem>
            <SelectItem value="key_disable">Key 禁用</SelectItem>
            <SelectItem value="budget_change">预算变更</SelectItem>
            <SelectItem value="budget_approve">预算审批</SelectItem>
            <SelectItem value="permission_change">权限变更</SelectItem>
            <SelectItem value="model_whitelist_change">白名单变更</SelectItem>
            <SelectItem value="org_structure_change">组织结构变更</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">操作记录</h3>
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground">时间</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">
                  操作类型
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">
                  操作人
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">
                  操作对象
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">详情</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <TableRow key={log.id} className="border-border/40 hover:bg-indigo-50/30">
                  <TableCell className="text-[12px] tabular-nums text-muted-foreground whitespace-nowrap">
                    {log.createdAt}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={`border-0 ${getActionStyle(log.action)}`}>
                      {actionLabels[log.action] ?? log.action}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-medium">{log.operator}</TableCell>
                  <TableCell className="text-sm">{log.target}</TableCell>
                  <TableCell className="text-sm text-muted-foreground max-w-64 truncate">
                    {log.detail}
                  </TableCell>
                  <TableCell className="font-mono text-[11px] text-muted-foreground">
                    {log.ip}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}

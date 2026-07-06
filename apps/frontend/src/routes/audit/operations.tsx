import { useState, useEffect } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { auditApi } from '@/api/audit'
import type { OperationLog } from '@/api/types'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

const PAGE_SIZE = 20

const actionLabels: Record<string, string> = {
  key_create: 'Key 创建',
  key_disable: 'Key 禁用',
  key_rotate: 'Key 轮转',
  budget_change: '预算变更',
  budget_approve: '预算审批',
  permission_change: '权限变更',
  role_assign: '角色分配',
  model_whitelist_change: '白名单变更',
  model_create: '模型创建',
  model_update: '模型修改',
  model_delete: '模型删除',
  alert_create: '预警创建',
  alert_update: '预警修改',
  alert_delete: '预警删除',
  member_add: '成员添加',
  member_remove: '成员移除',
  org_structure_change: '组织结构变更',
}

function getActionStyle(action: string): string {
  if (action.startsWith('key_')) return 'bg-amber-50 text-amber-700'
  if (action.startsWith('budget_')) return 'bg-emerald-50 text-emerald-700'
  if (action.startsWith('model_')) return 'bg-blue-50 text-blue-700'
  if (action.startsWith('alert_')) return 'bg-red-50 text-red-700'
  if (action.startsWith('permission') || action.startsWith('role'))
    return 'bg-primary/10 text-primary'
  return 'bg-slate-100 text-slate-600'
}

export default function OperationLogsPage() {
  const [logs, setLogs] = useState<OperationLog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [actionFilter, setActionFilter] = useState('all')
  const [timeRange, setTimeRange] = useState('all')

  useEffect(() => {
    const params: Record<string, string> = {
      page: String(page),
      pageSize: String(PAGE_SIZE),
    }
    if (actionFilter !== 'all') params.action = actionFilter
    if (timeRange !== 'all') params.timeRange = timeRange
    auditApi.getOperations(params).then(res => {
      setLogs(res.items)
      setTotal(res.total)
    })
  }, [page, actionFilter, timeRange])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="flex items-center gap-3">
        <Select value={timeRange} onValueChange={(v) => { setPage(1); setTimeRange(v) }}>
          <SelectTrigger className="w-36 h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部时间</SelectItem>
            <SelectItem value="today">今天</SelectItem>
            <SelectItem value="7d">最近 7 天</SelectItem>
            <SelectItem value="30d">最近 30 天</SelectItem>
          </SelectContent>
        </Select>
        <Select value={actionFilter} onValueChange={(v) => { setPage(1); setActionFilter(v) }}>
          <SelectTrigger className="w-36 h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部类型</SelectItem>
            <SelectItem value="key_create">Key 创建</SelectItem>
            <SelectItem value="key_disable">Key 禁用</SelectItem>
            <SelectItem value="key_rotate">Key 轮转</SelectItem>
            <SelectItem value="budget_change">预算变更</SelectItem>
            <SelectItem value="budget_approve">预算审批</SelectItem>
            <SelectItem value="permission_change">权限变更</SelectItem>
            <SelectItem value="role_assign">角色分配</SelectItem>
            <SelectItem value="model_create">模型创建</SelectItem>
            <SelectItem value="model_update">模型修改</SelectItem>
            <SelectItem value="model_delete">模型删除</SelectItem>
            <SelectItem value="alert_create">预警创建</SelectItem>
            <SelectItem value="alert_update">预警修改</SelectItem>
            <SelectItem value="alert_delete">预警删除</SelectItem>
            <SelectItem value="model_whitelist_change">白名单变更</SelectItem>
            <SelectItem value="member_add">成员添加</SelectItem>
            <SelectItem value="member_remove">成员移除</SelectItem>
            <SelectItem value="org_structure_change">组织结构变更</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border shadow-xs">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">时间</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">操作类型</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">操作人</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">操作对象</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">详情</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="py-8 text-center text-sm text-muted-foreground">
                  暂无操作记录
                </TableCell>
              </TableRow>
            ) : (
              logs.map((log) => (
                <TableRow key={log.id} className="even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="text-xs tabular-nums text-muted-foreground whitespace-nowrap">{log.createdAt}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className={cn('border-0', getActionStyle(log.action))}>
                      {actionLabels[log.action] ?? log.action}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm font-medium">{log.operator}</TableCell>
                  <TableCell className="text-sm">{log.target}</TableCell>
                  <TableCell className="text-sm text-muted-foreground max-w-64 truncate">{log.detail}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{log.ip}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        {/* Pagination */}
        <div className="flex items-center justify-between border-t border-border px-4 py-2.5">
          <span className="text-xs text-muted-foreground">
            共 <span className="tabular-nums font-medium text-foreground">{total}</span> 条记录
          </span>
          <div className="flex items-center gap-1.5">
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              aria-label="上一页"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              <ChevronLeft className="size-4" />
            </Button>
            <span className="text-xs tabular-nums text-muted-foreground">
              {page} / {totalPages}
            </span>
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              aria-label="下一页"
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
            >
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

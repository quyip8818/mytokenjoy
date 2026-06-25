import { ScrollText } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/ui/status-badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditToolbar } from '@/components/audit/audit-toolbar'
import { auditApi } from '@/api/audit'
import { useFilteredResource } from '@/hooks/use-filtered-resource'
import { getOperationActionBadgeVariant } from '@/lib/labels'
import { downloadCsv } from '@/lib/csv-export'

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

export default function OperationLogsPage() {
  const {
    data: logs = [],
    loading,
    filter: actionFilter,
    setFilter: setActionFilter,
  } = useFilteredResource(async (filter) => {
    const params = filter !== 'all' ? { action: filter } : undefined
    const res = await auditApi.getOperations(params)
    return res.items
  }, 'all')

  const handleExport = () => {
    downloadCsv(
      'operation-audit.csv',
      ['时间', '操作类型', '操作人', '操作对象', '详情', 'IP'],
      logs.map((log) => [
        log.createdAt,
        actionLabels[log.action] ?? log.action,
        log.operator,
        log.target,
        log.detail,
        log.ip,
      ]),
    )
  }

  return (
    <AuditFilteredPage
      title="操作记录"
      loading={loading}
      items={logs}
      empty={{
        icon: ScrollText,
        title: '暂无操作记录',
        description: '调整筛选条件或完成管理操作后，记录将显示在这里',
      }}
      actions={
        <div className="flex items-center gap-3">
          <Select value={actionFilter} onValueChange={(v) => setActionFilter(v ?? 'all')}>
            <SelectTrigger className="w-40 border-border/60 focus:ring-blue-500">
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
          <AuditToolbar onExport={handleExport} />
        </div>
      }
    >
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead>时间</TableHead>
            <TableHead>操作类型</TableHead>
            <TableHead>操作人</TableHead>
            <TableHead>操作对象</TableHead>
            <TableHead>详情</TableHead>
            <TableHead>IP</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {logs.map((log) => (
            <TableRow key={log.id}>
              <TableCell className="text-[12px] tabular-nums whitespace-nowrap text-muted-foreground">
                {log.createdAt}
              </TableCell>
              <TableCell>
                <StatusBadge variant={getOperationActionBadgeVariant(log.action)}>
                  {actionLabels[log.action] ?? log.action}
                </StatusBadge>
              </TableCell>
              <TableCell className="font-medium">{log.operator}</TableCell>
              <TableCell className="text-sm">{log.target}</TableCell>
              <TableCell className="max-w-64 truncate text-sm text-muted-foreground">
                {log.detail}
              </TableCell>
              <TableCell className="font-mono text-[11px] text-muted-foreground">
                {log.ip}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </AuditFilteredPage>
  )
}

import type { OperationLog } from '@/api/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/ui/status-badge'
import { getOperationActionBadgeVariant, OPERATION_ACTION_LABELS } from '@/features/audit'

interface OperationsLogTableProps {
  logs: OperationLog[]
}

export function OperationsLogTable({ logs }: OperationsLogTableProps) {
  return (
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
          <TableRow key={log.id} className="even:bg-muted/40">
            <TableCell className="text-[12px] tabular-nums whitespace-nowrap text-muted-foreground">
              {log.createdAt}
            </TableCell>
            <TableCell>
              <StatusBadge variant={getOperationActionBadgeVariant(log.action)}>
                {OPERATION_ACTION_LABELS[log.action] ?? log.action}
              </StatusBadge>
            </TableCell>
            <TableCell className="font-medium">{log.operator}</TableCell>
            <TableCell className="text-sm">{log.target}</TableCell>
            <TableCell className="max-w-64 truncate text-sm text-muted-foreground">
              {log.detail}
            </TableCell>
            <TableCell className="font-mono text-[11px] text-muted-foreground">{log.ip}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

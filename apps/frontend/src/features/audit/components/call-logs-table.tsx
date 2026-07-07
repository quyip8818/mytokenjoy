import { Fragment } from 'react'
import { ChevronRight } from 'lucide-react'
import type { CallLog } from '@/api/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/ui/status-badge'
import { CALL_LOG_STATUS_LABELS, CALL_LOG_STATUS_VARIANTS } from '@/features/audit'

export interface CallLogsTableProps {
  logs: readonly CallLog[]
  expandedId: string | null
  contentRetentionEnabled: boolean
  onToggleExpanded: (logId: string) => void
}

export function CallLogsTable({
  logs,
  expandedId,
  contentRetentionEnabled,
  onToggleExpanded,
}: CallLogsTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="w-6"></TableHead>
          <TableHead>时间</TableHead>
          <TableHead>调用方</TableHead>
          <TableHead>类型</TableHead>
          <TableHead>模型</TableHead>
          <TableHead className="text-right">输入 Token</TableHead>
          <TableHead className="text-right">输出 Token</TableHead>
          <TableHead className="text-right">延迟</TableHead>
          <TableHead className="text-right">费用 (¥)</TableHead>
          <TableHead>状态</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {logs.map((log) => {
          const statusVariant = CALL_LOG_STATUS_VARIANTS[log.status] ?? 'success'
          const isExpanded = expandedId === log.id
          const canExpand = contentRetentionEnabled
          return (
            <Fragment key={log.id}>
              <TableRow
                className={`even:bg-muted/40 ${canExpand ? 'cursor-pointer' : ''}`}
                onClick={() => onToggleExpanded(log.id)}
              >
                <TableCell className="w-6 pr-0">
                  {canExpand && (
                    <ChevronRight
                      className={`h-4 w-4 text-muted-foreground transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                    />
                  )}
                </TableCell>
                <TableCell className="text-[12px] tabular-nums whitespace-nowrap text-muted-foreground">
                  {log.createdAt}
                </TableCell>
                <TableCell className="font-medium">{log.caller}</TableCell>
                <TableCell>
                  <StatusBadge variant={log.callerType === 'platform_key' ? 'violet' : 'neutral'}>
                    {log.callerType === 'member' ? '成员' : '应用'}
                  </StatusBadge>
                </TableCell>
                <TableCell>
                  <StatusBadge variant="info">{log.model}</StatusBadge>
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {log.inputTokens.toLocaleString()}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {log.outputTokens.toLocaleString()}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">{log.latencyMs}ms</TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {log.cost.toFixed(2)}
                </TableCell>
                <TableCell>
                  <StatusBadge variant={statusVariant}>
                    {CALL_LOG_STATUS_LABELS[log.status] ?? log.status}
                  </StatusBadge>
                </TableCell>
              </TableRow>
              {isExpanded && canExpand && (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={10} className="bg-blue-50/20 p-4">
                    <div className="text-sm">
                      <div className="mb-1 text-xs font-medium text-muted-foreground">输入预览</div>
                      <div className="rounded-md border border-border/40 bg-background p-3 text-xs">
                        {log.previewSnippet || '—'}
                      </div>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </Fragment>
          )
        })}
      </TableBody>
    </Table>
  )
}

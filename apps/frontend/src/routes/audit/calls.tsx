import { Fragment, useState } from 'react'
import { Activity, ChevronRight } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditToolbar } from '@/components/audit/audit-toolbar'
import { StatusBadge } from '@/components/ui/status-badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { auditApi } from '@/api/audit'
import { useFilteredResource } from '@/hooks/use-filtered-resource'
import { useAuditSettings } from '@/hooks/use-audit-settings'
import { CALL_LOG_STATUS_VARIANTS } from '@/lib/labels'
import { downloadCsv } from '@/lib/csv-export'

const statusLabels: Record<string, string> = {
  success: '成功',
  error: '错误',
  filtered: '已过滤',
}

export default function CallLogsPage() {
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const { contentRetentionEnabled } = useAuditSettings()
  const {
    data: logs = [],
    loading,
    filter: statusFilter,
    setFilter: setStatusFilter,
  } = useFilteredResource(async (filter) => {
    const params = filter !== 'all' ? { status: filter } : undefined
    const res = await auditApi.getCalls(params)
    return res.items
  }, 'all')

  const handleExport = () => {
    downloadCsv(
      'call-audit.csv',
      ['时间', '调用方', '类型', '模型', '输入Token', '输出Token', '延迟', '费用', '状态'],
      logs.map((log) => [
        log.createdAt,
        log.caller,
        log.callerType === 'member' ? '成员' : '应用',
        log.model,
        log.inputTokens,
        log.outputTokens,
        `${log.latencyMs}ms`,
        log.cost.toFixed(2),
        statusLabels[log.status] ?? log.status,
      ]),
    )
  }

  return (
    <AuditFilteredPage
      title="调用记录"
      loading={loading}
      items={logs}
      empty={{
        icon: Activity,
        title: '暂无调用记录',
        description: '模型 API 调用成功后，日志将显示在这里',
      }}
      actions={
        <div className="flex items-center gap-3">
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v ?? 'all')}>
            <SelectTrigger className="w-32 border-border/60 focus:ring-blue-500">
              <SelectValue placeholder="全部状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="success">成功</SelectItem>
              <SelectItem value="error">错误</SelectItem>
              <SelectItem value="filtered">已过滤</SelectItem>
            </SelectContent>
          </Select>
          <AuditToolbar onExport={handleExport} />
        </div>
      }
    >
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
                  className={canExpand ? 'cursor-pointer' : undefined}
                  onClick={() => {
                    if (canExpand) setExpandedId(isExpanded ? null : log.id)
                  }}
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
                      {statusLabels[log.status] ?? log.status}
                    </StatusBadge>
                  </TableCell>
                </TableRow>
                {isExpanded && canExpand && (
                  <TableRow className="hover:bg-transparent">
                    <TableCell colSpan={10} className="bg-blue-50/20 p-4">
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <div className="mb-1 text-xs font-medium text-muted-foreground">
                            输入内容
                          </div>
                          <div className="rounded-md border border-border/40 bg-background p-3 text-xs">
                            {log.inputPreview}
                          </div>
                        </div>
                        <div>
                          <div className="mb-1 text-xs font-medium text-muted-foreground">
                            输出内容
                          </div>
                          <div className="rounded-md border border-border/40 bg-background p-3 text-xs">
                            {log.outputPreview}
                          </div>
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
    </AuditFilteredPage>
  )
}

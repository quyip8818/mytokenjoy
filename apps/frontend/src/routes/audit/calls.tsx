import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ChevronRight } from 'lucide-react'
import { auditApi } from '@/api/audit'
import type { CallLog } from '@/api/types'

const statusStyles: Record<string, { label: string; className: string }> = {
  success: { label: '成功', className: 'bg-emerald-50 text-emerald-700 border-0' },
  error: { label: '错误', className: 'bg-red-50 text-red-700 border-0' },
  filtered: { label: '已过滤', className: 'bg-amber-50 text-amber-700 border-0' },
}

export default function CallLogsPage() {
  const [logs, setLogs] = useState<CallLog[]>([])
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [expandedId, setExpandedId] = useState<string | null>(null)

  useEffect(() => {
    const params = statusFilter !== 'all' ? { status: statusFilter } : undefined
    auditApi.getCalls(params).then(res => setLogs(res.items))
  }, [statusFilter])

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-end">
        <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v ?? 'all')}>
          <SelectTrigger className="w-32 border-border/60 focus:ring-indigo-500">
            <SelectValue placeholder="全部状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="success">成功</SelectItem>
            <SelectItem value="error">错误</SelectItem>
            <SelectItem value="filtered">已过滤</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Card className="shadow-xs border-border">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">调用记录</h3>
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground w-6"></TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">时间</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">调用方</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">类型</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">模型</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">输入 Token</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">输出 Token</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">延迟</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">费用 (¥)</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => {
                const status = statusStyles[log.status] ?? statusStyles.success
                const isExpanded = expandedId === log.id
                return (
                  <>
                    <TableRow
                      key={log.id}
                      className="border-border-subtle cursor-pointer hover:bg-muted/50"
                      onClick={() => setExpandedId(isExpanded ? null : log.id)}
                    >
                      <TableCell className="w-6 pr-0">
                        <ChevronRight className={`h-4 w-4 text-muted-foreground transition-transform ${isExpanded ? 'rotate-90' : ''}`} />
                      </TableCell>
                      <TableCell className="text-[12px] tabular-nums text-muted-foreground whitespace-nowrap">{log.createdAt}</TableCell>
                      <TableCell className="font-medium">{log.caller}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className={`border-0 ${log.callerType === 'platform_key' ? 'bg-violet-50 text-violet-700' : ''}`}>
                          {log.callerType === 'member' ? '成员' : '应用'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary" className="bg-primary/10 text-primary border-0">{log.model}</Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono text-xs">{log.inputTokens.toLocaleString()}</TableCell>
                      <TableCell className="text-right font-mono text-xs">{log.outputTokens.toLocaleString()}</TableCell>
                      <TableCell className="text-right font-mono text-xs">{log.latencyMs}ms</TableCell>
                      <TableCell className="text-right font-mono text-xs">{log.cost.toFixed(2)}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className={status.className}>{status.label}</Badge>
                      </TableCell>
                    </TableRow>
                    {isExpanded && (
                      <TableRow key={`${log.id}-detail`} className="border-border-subtle hover:bg-transparent">
                        <TableCell colSpan={10} className="bg-muted/30 p-4">
                          <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                              <div className="text-xs font-medium text-muted-foreground mb-1">输入内容</div>
                              <div className="bg-background rounded-md p-3 border border-border/40 text-xs">{log.inputPreview}</div>
                            </div>
                            <div>
                              <div className="text-xs font-medium text-muted-foreground mb-1">输出内容</div>
                              <div className="bg-background rounded-md p-3 border border-border/40 text-xs">{log.outputPreview}</div>
                            </div>
                          </div>
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                )
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}

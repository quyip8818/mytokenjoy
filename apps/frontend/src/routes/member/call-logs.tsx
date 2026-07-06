import { useState } from 'react'
import { Link } from 'react-router'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { mockCallLogs } from '@/mocks/data'
import { ChevronLeft, ChevronRight, ArrowLeft } from 'lucide-react'

const CURRENT_USER_ID = 'm-1'
const PAGE_SIZE = 20

const allLogs = mockCallLogs.filter((l) => l.callerId === CURRENT_USER_ID)

export default function MemberCallLogsPage() {
  const [page, setPage] = useState(1)
  const totalPages = Math.max(1, Math.ceil(allLogs.length / PAGE_SIZE))
  const pagedLogs = allLogs.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/me" className="text-xs text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
        </Link>
        <h1 className="text-sm font-semibold">使用记录</h1>
      </div>

      <div className="rounded-lg border border-border bg-card shadow-xs">
        <div className="divide-y divide-border">
          {pagedLogs.map((log) => (
            <div key={log.id} className="flex items-center gap-4 px-5 py-3">
              <span className="text-xs tabular-nums text-muted-foreground w-28 shrink-0">
                {log.createdAt.slice(5, 16)}
              </span>
              <span className="text-sm font-medium w-40 truncate">{log.model}</span>
              <span className="text-xs tabular-nums text-muted-foreground">
                {log.inputTokens.toLocaleString()} + {log.outputTokens.toLocaleString()} tok
              </span>
              <span className="text-xs tabular-nums text-muted-foreground ml-auto">
                ¥{log.cost.toFixed(2)}
              </span>
              <Badge variant="outline" className={cn(
                'text-xs',
                log.status === 'success' ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                  : log.status === 'error' ? 'bg-red-50 text-red-700 border-red-200'
                  : 'bg-amber-50 text-amber-700 border-amber-200'
              )}>
                {log.status === 'success' ? '成功' : log.status === 'error' ? '失败' : '过滤'}
              </Badge>
            </div>
          ))}
        </div>

        <div className="flex items-center justify-between border-t border-border px-5 py-2.5">
          <span className="text-xs text-muted-foreground">
            共 <span className="tabular-nums font-medium text-foreground">{allLogs.length}</span> 条
          </span>
          <div className="flex items-center gap-1.5">
            <Button variant="ghost" size="icon" className="size-7" aria-label="上一页" disabled={page <= 1} onClick={() => setPage(page - 1)}>
              <ChevronLeft className="size-4" />
            </Button>
            <span className="text-xs tabular-nums text-muted-foreground">{page} / {totalPages}</span>
            <Button variant="ghost" size="icon" className="size-7" aria-label="下一页" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

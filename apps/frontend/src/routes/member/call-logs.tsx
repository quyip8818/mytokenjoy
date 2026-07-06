import { Link } from 'react-router'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ArrowLeft } from 'lucide-react'
import { useMemberCallLogsPage } from '@/routes/member/hooks/use-member-call-logs-page'

export default function MemberCallLogsPage() {
  const { logs, total, page, totalPages, loading, setPage } = useMemberCallLogsPage()

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/me" className="text-xs text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
        </Link>
        <h1 className="text-sm font-semibold">使用记录</h1>
      </div>

      <div className="rounded-lg border border-border bg-card shadow-xs">
        {loading ? (
          <p className="px-5 py-8 text-center text-sm text-muted-foreground">加载中…</p>
        ) : (
          <div className="divide-y divide-border">
            {logs.map((log) => (
              <div key={log.id} className="flex items-center gap-4 px-5 py-3">
                <span className="w-28 shrink-0 text-xs tabular-nums text-muted-foreground">
                  {log.createdAt.slice(5, 16)}
                </span>
                <span className="w-40 truncate text-sm font-medium">{log.model}</span>
                <span className="text-xs tabular-nums text-muted-foreground">
                  {log.inputTokens.toLocaleString()} + {log.outputTokens.toLocaleString()} tok
                </span>
                <span className="ml-auto text-xs tabular-nums text-muted-foreground">
                  ¥{log.cost.toFixed(2)}
                </span>
                <Badge
                  variant="outline"
                  className={cn(
                    'text-xs',
                    log.status === 'success'
                      ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                      : log.status === 'error'
                        ? 'border-red-200 bg-red-50 text-red-700'
                        : 'border-amber-200 bg-amber-50 text-amber-700',
                  )}
                >
                  {log.status === 'success' ? '成功' : log.status === 'error' ? '失败' : '过滤'}
                </Badge>
              </div>
            ))}
          </div>
        )}

        <div className="flex items-center justify-between border-t border-border px-5 py-2.5">
          <span className="text-xs text-muted-foreground">
            共 <span className="font-medium tabular-nums text-foreground">{total}</span> 条
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

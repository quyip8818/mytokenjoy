import { cn } from '@/lib/utils'

export interface PaginationProps {
  page: number
  totalPages: number
  total: number
  onChange: (page: number) => void
  className?: string
}

export function Pagination({ page, totalPages, total, onChange, className }: PaginationProps) {
  if (total === 0) return null

  return (
    <div
      className={cn('flex items-center justify-between text-sm text-muted-foreground', className)}
    >
      <span>共 {total} 条</span>
      <div className="flex items-center gap-1">
        <button
          disabled={page <= 1}
          onClick={() => onChange(page - 1)}
          className="rounded border px-3 py-1 disabled:opacity-40"
        >
          上一页
        </button>
        <span className="px-3">
          {page} / {totalPages}
        </span>
        <button
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
          className="rounded border px-3 py-1 disabled:opacity-40"
        >
          下一页
        </button>
      </div>
    </div>
  )
}

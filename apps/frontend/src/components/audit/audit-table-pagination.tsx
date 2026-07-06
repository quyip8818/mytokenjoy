import { Button } from '@/components/ui/button'

interface AuditTablePaginationProps {
  total: number
  page: number
  totalPages: number
  onPageChange: (page: number) => void
}

export function AuditTablePagination({
  total,
  page,
  totalPages,
  onPageChange,
}: AuditTablePaginationProps) {
  if (totalPages <= 1) {
    return null
  }

  return (
    <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
      <span>共 {total} 条</span>
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size="sm"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          上一页
        </Button>
        <span className="px-3 py-1">
          {page} / {totalPages}
        </span>
        <Button
          variant="outline"
          size="sm"
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          下一页
        </Button>
      </div>
    </div>
  )
}

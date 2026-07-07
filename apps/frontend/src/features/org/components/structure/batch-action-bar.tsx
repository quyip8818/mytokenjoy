import { Button } from '@/components/ui/button'
import { ArrowRightLeft, CheckCircle2, XCircle, Trash2, X } from 'lucide-react'

interface BatchActionBarProps {
  count: number
  onTransfer: () => void
  onEnable: () => void
  onDisable: () => void
  onDelete: () => void
  onClear: () => void
}

export function BatchActionBar({
  count,
  onTransfer,
  onEnable,
  onDisable,
  onDelete,
  onClear,
}: BatchActionBarProps) {
  if (count === 0) return null

  return (
    <div className="absolute bottom-4 left-1/2 z-50 -translate-x-1/2 animate-in slide-in-from-bottom-4 fade-in-0 duration-200">
      <div className="flex items-center gap-2 rounded-lg border border-border bg-card px-4 py-2.5 shadow-lg">
        <span className="text-sm font-medium text-foreground">
          已选 <span className="font-semibold tabular-nums">{count}</span> 人
        </span>
        <div className="mx-2 h-4 w-px bg-border" />
        <Button variant="ghost" size="sm" onClick={onTransfer}>
          <ArrowRightLeft className="size-3.5" />
          转移部门
        </Button>
        <Button variant="ghost" size="sm" onClick={onEnable}>
          <CheckCircle2 className="size-3.5" />
          启用
        </Button>
        <Button variant="ghost" size="sm" onClick={onDisable}>
          <XCircle className="size-3.5" />
          停用
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive hover:bg-red-50"
          onClick={onDelete}
        >
          <Trash2 className="size-3.5" />
          删除
        </Button>
        <div className="mx-1 h-4 w-px bg-border" />
        <Button variant="ghost" size="icon-xs" onClick={onClear} aria-label="取消选择">
          <X className="size-3.5" />
        </Button>
      </div>
    </div>
  )
}

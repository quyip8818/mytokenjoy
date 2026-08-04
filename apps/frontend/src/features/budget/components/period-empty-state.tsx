import { ClipboardList } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface PeriodEmptyStateProps {
  onInherit: () => void
  onStartFresh: () => void
  loading?: boolean
}

export function PeriodEmptyState({ onInherit, onStartFresh, loading }: PeriodEmptyStateProps) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
      <ClipboardList className="size-10 text-muted-foreground" strokeWidth={1.5} />
      <div>
        <p className="text-sm font-medium text-foreground">下月尚未配置预算</p>
        <p className="mt-1 text-xs text-muted-foreground">可从本月配置继承，或从零开始设置。</p>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Button onClick={onInherit} disabled={loading}>
          {loading ? '继承中…' : '继承本月配置'}
        </Button>
        <button
          type="button"
          className="text-xs text-muted-foreground hover:text-foreground"
          onClick={onStartFresh}
        >
          从零开始
        </button>
      </div>
    </div>
  )
}

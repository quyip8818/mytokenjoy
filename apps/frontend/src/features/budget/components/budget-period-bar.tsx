import { ChevronLeft, ChevronRight, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { getCurrentBudgetPeriod } from '@/lib/date'
import { formatBudgetPeriodLabel, shiftBudgetPeriod } from '../lib/mappers'
import { getPeriodState, type PeriodState } from '../lib/period-state'

interface BudgetPeriodBarProps {
  period: string
  onShiftPeriod: (delta: number) => void
  onPreConfigNext: () => void
  /** Earliest period that has data (for left arrow boundary) */
  earliestPeriod?: string
}

const BADGE_CONFIG: Record<PeriodState, { label: string; className: string }> = {
  past: { label: '已结束', className: 'bg-slate-100 text-slate-600 border-0' },
  current: { label: '本月', className: 'bg-blue-50 text-blue-700 border-0' },
  future: { label: '预配中', className: 'bg-amber-50 text-amber-700 border-0' },
}

export function BudgetPeriodBar({
  period,
  onShiftPeriod,
  onPreConfigNext,
  earliestPeriod,
}: BudgetPeriodBarProps) {
  const currentPeriod = getCurrentBudgetPeriod()
  const maxPeriod = shiftBudgetPeriod(currentPeriod, 1)
  const state = getPeriodState(period)

  const canGoLeft = !earliestPeriod || period > earliestPeriod
  const canGoRight = period < maxPeriod

  return (
    <div className="flex h-11 items-center justify-between border-b border-border px-5">
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          onClick={() => onShiftPeriod(-1)}
          disabled={!canGoLeft}
          aria-label="上一月"
        >
          <ChevronLeft className="size-4" />
        </Button>

        <span className="text-sm font-semibold text-foreground">
          {formatBudgetPeriodLabel(period)}
        </span>

        <Badge variant="outline" className={cn('h-5 text-xs', BADGE_CONFIG[state].className)}>
          {BADGE_CONFIG[state].label}
        </Badge>

        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          onClick={() => onShiftPeriod(1)}
          disabled={!canGoRight}
          aria-label="下一月"
        >
          <ChevronRight className="size-4" />
        </Button>
      </div>

      {state === 'current' && (
        <Button
          variant="ghost"
          className="h-7 gap-1.5 text-xs text-muted-foreground"
          onClick={onPreConfigNext}
        >
          预配下月
          <ArrowRight className="size-3.5" />
        </Button>
      )}
    </div>
  )
}

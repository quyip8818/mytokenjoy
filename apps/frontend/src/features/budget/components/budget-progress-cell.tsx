import { Progress } from '@/components/ui/progress'
import { getBudgetProgressClass } from '@/features/budget'
import { cn } from '@/lib/utils'

interface BudgetProgressCellProps {
  value: number
  total: number
  className?: string
  labelClassName?: string
  accentLabel?: boolean
  showPercent?: boolean
}

export function BudgetProgressCell({
  value,
  total,
  className,
  labelClassName,
  accentLabel = false,
  showPercent = true,
}: BudgetProgressCellProps) {
  const pct = total > 0 ? Math.round((value / total) * 100) : 0

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Progress value={pct} className="h-2 flex-1" />
      {showPercent && (
        <span
          className={cn(
            'text-xs',
            accentLabel && 'font-semibold',
            getBudgetProgressClass(pct, accentLabel),
            labelClassName,
          )}
        >
          {pct}%
        </span>
      )}
    </div>
  )
}

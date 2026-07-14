import { Progress } from '@/components/ui/progress'
import { getBudgetProgressClass, getBudgetProgressTone } from '@/features/budget'
import { cn } from '@/lib/utils'
import { Link } from 'react-router'
import { ArrowRight } from 'lucide-react'
import { formatDisplayCurrency } from '@/lib/points'

interface BudgetHeroCardProps {
  budget: number
  consumed: number
  loading: boolean
}

export function BudgetHeroCard({ budget, consumed, loading }: BudgetHeroCardProps) {
  const pct = budget > 0 ? Math.round((consumed / budget) * 100) : 0
  const tone = getBudgetProgressTone(pct)

  if (loading) {
    return (
      <div className="rounded-lg border bg-card p-5">
        <div className="h-5 w-32 animate-pulse rounded bg-muted" />
        <div className="mt-3 h-4 w-full animate-pulse rounded bg-muted" />
      </div>
    )
  }

  if (budget <= 0) {
    return (
      <div className="rounded-lg border bg-card p-5">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">本月预算</span>
          <Link
            to="/budget"
            className="flex items-center gap-1 text-xs text-muted-foreground hover:text-primary"
          >
            设置预算
            <ArrowRight className="size-3" />
          </Link>
        </div>
        <p className="mt-2 text-sm text-muted-foreground">暂未设置预算</p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border bg-card p-5">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">预算消耗</span>
        <Link
          to="/budget"
          className="flex items-center gap-1 text-xs text-muted-foreground hover:text-primary"
        >
          查看预算详情
          <ArrowRight className="size-3" />
        </Link>
      </div>
      <div className="mt-3 flex items-end gap-3">
        <span className="text-2xl font-bold tabular-nums">{formatDisplayCurrency(consumed)}</span>
        <span className="mb-0.5 text-sm text-muted-foreground">
          / {formatDisplayCurrency(budget)}
        </span>
      </div>
      <div className="mt-3 flex items-center gap-3">
        <Progress
          value={pct}
          className={cn(
            'h-3 flex-1',
            tone === 'danger' && '[&>div]:bg-red-500',
            tone === 'warning' && '[&>div]:bg-amber-500',
          )}
        />
        <span
          className={cn('text-sm font-semibold tabular-nums', getBudgetProgressClass(pct, true))}
        >
          {pct}%
        </span>
      </div>
    </div>
  )
}

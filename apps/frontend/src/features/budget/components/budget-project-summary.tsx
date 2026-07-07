import type { BudgetProjectView } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { POLICY_LABELS } from '@/features/budget/lib/constants'

type BudgetProjectHeaderProps = {
  project: BudgetProjectView
}

export function BudgetProjectHeader({ project }: BudgetProjectHeaderProps) {
  const policy = POLICY_LABELS[project.overrunPolicy]

  return (
    <div className="flex items-center gap-3">
      <h3 className="text-sm font-semibold text-foreground">{project.name}</h3>
      <Badge variant="outline" className="text-xs font-normal">
        所属：{project.departmentName}
      </Badge>
      <Badge variant="outline" className={cn(policy.className, 'text-xs')}>
        {policy.label}
      </Badge>
    </div>
  )
}

type BudgetProjectSummaryProps = {
  project: BudgetProjectView
}

function SummaryCard({
  label,
  value,
  muted,
  highlight,
}: {
  label: string
  value: number
  muted?: boolean
  highlight?: boolean
}) {
  return (
    <div className="rounded-lg border border-border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p
        className={cn(
          'mt-1 text-lg font-semibold tabular-nums',
          highlight ? 'text-red-600' : muted ? 'text-muted-foreground' : 'text-foreground',
        )}
      >
        ¥{value.toLocaleString()}
      </p>
    </div>
  )
}

export function BudgetProjectSummary({ project }: BudgetProjectSummaryProps) {
  const remaining = project.budget - project.consumed

  return (
    <div className="grid grid-cols-3 gap-4">
      <SummaryCard label="项目额度" value={project.budget} />
      <SummaryCard label="已消耗" value={project.consumed} muted />
      <SummaryCard label="剩余" value={remaining} highlight={remaining < 0} />
    </div>
  )
}

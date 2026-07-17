import { Bell, BellRing, Building2, FolderKanban } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface AlertStats {
  total: number
  enabled: number
  teamCoverage: { covered: number; total: number }
  projectCoverage: { covered: number; total: number }
}

interface StatCardProps {
  icon: React.ElementType
  label: string
  value: string | number
  sub?: string
  className?: string
}

function StatCard({ icon: Icon, label, value, sub, className }: StatCardProps) {
  return (
    <div
      className={cn(
        'flex items-center gap-3 rounded-xl border border-border bg-card p-4',
        className,
      )}
    >
      <span className="flex size-9 items-center justify-center rounded-lg bg-primary/8 text-primary">
        <Icon className="size-4" strokeWidth={1.75} />
      </span>
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-lg font-semibold tabular-nums text-foreground">
          {value}
          {sub && <span className="ml-1 text-xs font-normal text-muted-foreground">{sub}</span>}
        </p>
      </div>
    </div>
  )
}

interface BudgetAlertsStatsProps {
  stats: AlertStats
}

export function BudgetAlertsStats({ stats }: BudgetAlertsStatsProps) {
  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard icon={Bell} label="规则总数" value={stats.total} />
      <StatCard
        icon={BellRing}
        label="已启用"
        value={stats.enabled}
        sub={stats.total > 0 ? `/ ${stats.total}` : undefined}
      />
      <StatCard
        icon={Building2}
        label="团队覆盖"
        value={stats.teamCoverage.covered}
        sub={`/ ${stats.teamCoverage.total}`}
      />
      <StatCard
        icon={FolderKanban}
        label="项目覆盖"
        value={stats.projectCoverage.covered}
        sub={`/ ${stats.projectCoverage.total}`}
      />
    </div>
  )
}

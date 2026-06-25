import { StatCard } from '@/components/ui/stat-card'
import type { CostStatItem } from '@/routes/dashboard/hooks/use-cost-dashboard'

interface CostSummaryStatsProps {
  stats: CostStatItem[]
  loading: boolean
}

export function CostSummaryStats({ stats, loading }: CostSummaryStatsProps) {
  return (
    <div className="grid grid-cols-2 gap-5 lg:grid-cols-6">
      {stats.map((stat) => (
        <StatCard
          key={stat.label}
          label={stat.label}
          value={loading ? '-' : stat.value}
          icon={stat.icon}
          iconAccent={stat.accent}
        />
      ))}
    </div>
  )
}

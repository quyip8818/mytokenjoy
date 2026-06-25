import { StatCard } from '@/components/ui/stat-card'
import { formatMom, type CostStatItem } from '@/lib/dashboard'

interface CostSummaryStatsProps {
  stats: CostStatItem[]
  loading: boolean
}

export function CostSummaryStats({ stats, loading }: CostSummaryStatsProps) {
  return (
    <div className="grid grid-cols-2 gap-5 lg:grid-cols-5">
      {stats.map((stat) => (
        <StatCard
          key={stat.label}
          label={stat.label}
          value={loading ? '-' : stat.value}
          subValue={
            !loading && stat.mom !== undefined ? `环比 ${formatMom(stat.mom)}` : undefined
          }
          icon={stat.icon}
          iconAccent={stat.accent}
        />
      ))}
    </div>
  )
}

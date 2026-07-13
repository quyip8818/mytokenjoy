import { StatCard } from '@/components/ui/stat-card'
import { Activity, AlertTriangle, Timer } from 'lucide-react'
import type { CallsSummary } from '@/api/types'

interface CallLogsSummaryCardsProps {
  summary: CallsSummary | null
  loading: boolean
}

export function CallLogsSummaryCards({ summary, loading }: CallLogsSummaryCardsProps) {
  return (
    <div className="grid grid-cols-3 gap-4">
      <StatCard
        label="总调用数"
        value={loading ? '-' : (summary?.totalCalls.toLocaleString() ?? '0')}
        icon={Activity}
        iconAccent="bg-primary"
        iconAccentStyle="solid"
      />
      <StatCard
        label="错误率"
        value={loading ? '-' : `${(summary?.errorRate ?? 0).toFixed(1)}%`}
        icon={AlertTriangle}
        iconAccent={(summary?.errorRate ?? 0) > 5 ? 'bg-red-500' : 'bg-emerald-500'}
        iconAccentStyle="solid"
      />
      <StatCard
        label="平均延迟"
        value={loading ? '-' : `${Math.round(summary?.avgLatencyMs ?? 0)}ms`}
        icon={Timer}
        iconAccent="bg-amber-400"
        iconAccentStyle="solid"
      />
    </div>
  )
}

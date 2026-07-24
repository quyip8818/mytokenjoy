import { Zap, DollarSign, Activity, Send, BarChart3, Coins, Gauge, Clock } from 'lucide-react'
import type {
  AccountStats,
  PerformanceStats,
  ResourceConsumption,
  UsageStats,
} from '@/api/types/mydashboard'
import { MyStatGroup } from '@/features/mydashboard'
import { formatDisplayCurrency, formatMoney } from '@/lib/quota-display'

interface MyDashboardStatsProps {
  loading: boolean
  accountData: AccountStats
  usageStats: UsageStats
  resourceConsumption: ResourceConsumption
  performance: PerformanceStats
}

export function MyDashboardStats({
  loading,
  accountData,
  usageStats,
  resourceConsumption,
  performance,
}: MyDashboardStatsProps) {
  return (
    <div className="grid grid-cols-4 gap-4">
      <MyStatGroup
        title="账户数据"
        icon={Coins}
        items={[
          {
            label: '预算剩余',
            value: loading ? '—' : formatDisplayCurrency(accountData.budgetRemaining),
            icon: Coins,
          },
          {
            label: '历史消耗',
            value: loading ? '—' : formatMoney(accountData.totalSpent),
            icon: DollarSign,
          },
        ]}
      />
      <MyStatGroup
        title="使用统计"
        icon={Zap}
        items={[
          {
            label: '请求次数',
            value: loading ? '—' : String(usageStats.requestCount),
            icon: Send,
          },
          {
            label: '统计次数',
            value: loading ? '—' : String(usageStats.totalCount),
            icon: BarChart3,
          },
        ]}
      />
      <MyStatGroup
        title="资源消耗"
        icon={DollarSign}
        items={[
          {
            label: '统计额度',
            value: loading ? '—' : formatMoney(resourceConsumption.totalCost),
            icon: Coins,
          },
          {
            label: '统计 Tokens',
            value: loading ? '—' : String(resourceConsumption.totalTokens),
            icon: Activity,
          },
        ]}
      />
      <MyStatGroup
        title="性能指标"
        icon={Activity}
        items={[
          {
            label: '平均 RPM',
            value: loading ? '—' : performance.avgRPM.toFixed(3),
            icon: Gauge,
          },
          {
            label: '平均 TPM',
            value: loading ? '—' : String(performance.avgTPM),
            icon: Clock,
          },
        ]}
      />
    </div>
  )
}

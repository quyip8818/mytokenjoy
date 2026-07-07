import {
  Wallet,
  Zap,
  DollarSign,
  Activity,
  Send,
  BarChart3,
  Coins,
  Gauge,
  Clock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import type {
  AccountStats,
  PerformanceStats,
  ResourceConsumption,
  UsageStats,
} from '@/api/types/member'
import { MemberStatGroup } from '@/features/member'

interface MemberDashboardStatsProps {
  loading: boolean
  accountData: AccountStats
  usageStats: UsageStats
  resourceConsumption: ResourceConsumption
  performance: PerformanceStats
  onRecharge: () => void
}

export function MemberDashboardStats({
  loading,
  accountData,
  usageStats,
  resourceConsumption,
  performance,
  onRecharge,
}: MemberDashboardStatsProps) {
  return (
    <div className="grid grid-cols-4 gap-4">
      <MemberStatGroup
        title="账户数据"
        icon={Wallet}
        items={[
          {
            label: '当前余额',
            value: loading ? '—' : `¥${accountData.balance.toFixed(2)}`,
            icon: Coins,
            action: (
              <Button variant="outline" size="sm" className="h-6 px-2 text-xs" onClick={onRecharge}>
                充值
              </Button>
            ),
          },
          {
            label: '历史消耗',
            value: loading ? '—' : `¥${accountData.totalSpent.toFixed(2)}`,
            icon: DollarSign,
          },
        ]}
      />
      <MemberStatGroup
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
      <MemberStatGroup
        title="资源消耗"
        icon={DollarSign}
        items={[
          {
            label: '统计额度',
            value: loading ? '—' : `¥${resourceConsumption.totalCost.toFixed(2)}`,
            icon: Coins,
          },
          {
            label: '统计 Tokens',
            value: loading ? '—' : String(resourceConsumption.totalTokens),
            icon: Activity,
          },
        ]}
      />
      <MemberStatGroup
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

import { Wallet, TrendingUp, BarChart3 } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'

interface WalletStatsProps {
  loading: boolean
  balance: number
  totalConsumed: number
  totalRequests: number
}

export function WalletStats({ loading, balance, totalConsumed, totalRequests }: WalletStatsProps) {
  return (
    <div className="grid grid-cols-3 gap-4">
      <StatCard icon={Wallet} label="当前余额" value={loading ? '—' : `¥${balance.toFixed(2)}`} />
      <StatCard
        icon={TrendingUp}
        label="历史消耗"
        value={loading ? '—' : `¥${totalConsumed.toFixed(2)}`}
      />
      <StatCard icon={BarChart3} label="请求次数" value={loading ? '—' : String(totalRequests)} />
    </div>
  )
}

import { AlertTriangle, BarChart3, Gift, TrendingUp, Wallet } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'
import type { WalletView } from '@/api/billing'
import { primaryWalletBalance } from '../lib/selectors'

interface WalletStatsProps {
  wallet: WalletView | undefined
  loading: boolean
}

export function WalletStats({ wallet, loading }: WalletStatsProps) {
  const primary = primaryWalletBalance(wallet)
  const balance = primary?.balance ?? 0
  const totalConsumed = primary?.totalConsumed ?? 0
  const totalTopup = primary?.totalTopup ?? 0
  const walletRemainPoint = wallet?.walletRemainPoint ?? 0
  const giftPoints = wallet?.giftPoints ?? 0
  const overdraftPoints = wallet?.overdraftPoints ?? 0
  const totalRequests = wallet?.totalRequests ?? 0
  const balances = wallet?.balances ?? []
  const hasOverdraft = overdraftPoints > 0

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          icon={Wallet}
          iconLayout="inline"
          label="当前余额"
          value={loading ? '—' : `¥${balance.toFixed(2)}`}
        />
        <StatCard
          icon={TrendingUp}
          iconLayout="inline"
          label="历史消耗"
          value={loading ? '—' : `¥${totalConsumed.toFixed(2)}`}
        />
        <StatCard
          icon={BarChart3}
          iconLayout="inline"
          label="可用 point"
          value={loading ? '—' : walletRemainPoint.toLocaleString()}
        />
        <StatCard
          icon={Gift}
          iconLayout="inline"
          label="赠送 point"
          value={loading ? '—' : giftPoints.toLocaleString()}
        />
      </div>

      {balances.length > 1 && (
        <div className="rounded-lg border border-border p-3">
          <p className="mb-2 text-xs font-medium text-muted-foreground">按币种余额</p>
          <div className="space-y-1">
            {balances.map((entry) => (
              <div
                key={entry.currency}
                className="flex items-center justify-between text-sm tabular-nums"
              >
                <span className="text-muted-foreground">{entry.currency}</span>
                <span>
                  ¥{entry.balance.toFixed(2)} / 充值 ¥{entry.totalTopup.toFixed(2)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {!loading && hasOverdraft && (
        <div className="flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          <AlertTriangle className="size-4 shrink-0" />
          <span>
            透支额度 {overdraftPoints.toLocaleString()} point，请尽快充值（累计充值 ¥
            {totalTopup.toFixed(2)}）
          </span>
        </div>
      )}

      <p className="text-xs text-muted-foreground">
        累计请求 {loading ? '—' : totalRequests.toLocaleString()} 次
      </p>
    </div>
  )
}

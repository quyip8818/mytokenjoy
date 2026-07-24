import { AlertTriangle, BarChart3, Gift, TrendingUp, Wallet } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'
import type { WalletView } from '@/api/billing'
import { formatCurrencyAmount, formatMoney } from '@/lib/quota-display'
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
  const walletRemainQuota = wallet?.walletRemainQuota ?? 0
  const giftQuota = wallet?.giftQuota ?? 0
  const overdraftQuota = wallet?.overdraftQuota ?? 0
  const totalRequests = wallet?.totalRequests ?? 0
  const balances = wallet?.balances ?? []
  const hasOverdraft = overdraftQuota > 0

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          icon={Wallet}
          iconLayout="inline"
          label="当前余额"
          value={loading ? '—' : formatMoney(balance)}
        />
        <StatCard
          icon={TrendingUp}
          iconLayout="inline"
          label="历史消耗"
          value={loading ? '—' : formatMoney(totalConsumed)}
        />
        <StatCard
          icon={BarChart3}
          iconLayout="inline"
          label="可用 quota"
          value={loading ? '—' : walletRemainQuota.toLocaleString()}
        />
        <StatCard
          icon={Gift}
          iconLayout="inline"
          label="赠送 quota"
          value={loading ? '—' : giftQuota.toLocaleString()}
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
                  {formatCurrencyAmount(entry.balance, entry.currency)} / 充值{' '}
                  {formatCurrencyAmount(entry.totalTopup, entry.currency)}
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
            透支额度 {overdraftQuota.toLocaleString()} quota，请尽快充值（累计充值{' '}
            {formatMoney(totalTopup)}）
          </span>
        </div>
      )}

      <p className="text-xs text-muted-foreground">
        累计请求 {loading ? '—' : totalRequests.toLocaleString()} 次
      </p>
    </div>
  )
}

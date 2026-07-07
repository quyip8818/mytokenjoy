import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useWalletPage } from '@/features/wallet'
import { RechargePanel } from './recharge-panel'
import { RechargeRecordsTable } from './recharge-records-table'
import { WalletStats } from './wallet-stats'

type WalletPageShellProps = ReturnType<typeof useWalletPage>

export function WalletPageShell({
  balance,
  currency,
  loading,
  error,
  refresh,
  topUpRecords,
  rechargePending,
  handleRecharge,
  totalConsumed,
  totalRequests,
}: WalletPageShellProps) {
  return (
    <PageShell
      description={
        <div>
          <h1 className="text-xl font-semibold">钱包管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">账户余额与充值管理</p>
        </div>
      }
      stats={
        <WalletStats
          loading={loading}
          balance={balance}
          totalConsumed={totalConsumed}
          totalRequests={totalRequests}
        />
      }
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6 p-0 pt-0"
        className="border-0 shadow-none"
      >
        <RechargePanel
          currency={currency}
          rechargePending={rechargePending}
          onRecharge={handleRecharge}
        />
        <RechargeRecordsTable records={topUpRecords} />
      </DataSection>
    </PageShell>
  )
}

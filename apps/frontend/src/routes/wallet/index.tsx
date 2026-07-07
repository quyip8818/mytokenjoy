import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { RechargePanel, RechargeRecordsTable, useWalletPage, WalletStats } from '@/features/wallet'

export default function WalletPage() {
  const {
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
  } = useWalletPage()

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

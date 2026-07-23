import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useWalletPage } from '@/features/wallet'
import { useSession } from '@/features/session'
import { walletBillingCurrency } from '../lib/selectors'
import { RechargePanel } from './recharge-panel'
import { RechargeRecordsTable } from './recharge-records-table'
import { WalletStats } from './wallet-stats'

type WalletPageShellProps = ReturnType<typeof useWalletPage>

export function WalletPageShell({
  wallet,
  loading,
  error,
  refresh,
  topUpRecords,
  rechargePending,
  handleRecharge,
}: WalletPageShellProps) {
  const { companyType } = useSession()
  const canRecharge = companyType !== 'trial' && companyType !== 'demo'

  return (
    <PageShell
      description={
        <div>
          <h1 className="text-xl font-semibold">钱包管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">账户余额与充值管理</p>
        </div>
      }
      stats={<WalletStats wallet={wallet} loading={loading} />}
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6 p-0 pt-0"
        className="border-0 shadow-none"
      >
        {canRecharge ? (
          <RechargePanel
            currency={walletBillingCurrency(wallet)}
            rechargePending={rechargePending}
            onRecharge={handleRecharge}
          />
        ) : (
          <div className="rounded-lg border border-border bg-card p-5">
            <p className="text-sm text-muted-foreground">
              试用账户不支持充值，升级为正式版后可使用充值功能。
            </p>
          </div>
        )}
        <RechargeRecordsTable records={topUpRecords} />
      </DataSection>
    </PageShell>
  )
}

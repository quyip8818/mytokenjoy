import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { useBillingPage } from '@/features/billing'
import { useSession } from '@/features/session'
import { walletBillingCurrency } from '../lib/selectors'
import { RechargePanel } from './recharge-panel'
import { RechargeRecordsTable } from './recharge-records-table'
import { BillingStats } from './billing-stats'

type BillingPageShellProps = ReturnType<typeof useBillingPage>

export function BillingPageShell({
  wallet,
  loading,
  error,
  refresh,
  topUpRecords,
  rechargePending,
  handleRecharge,
}: BillingPageShellProps) {
  const { companyType } = useSession()
  const canRecharge = companyType !== 'trial' && companyType !== 'demo'

  return (
    <PageShell>
      <PageHeader title="钱包管理" description="账户余额与充值管理" />

      <BillingStats wallet={wallet} loading={loading} />

      <DataSection loading={loading} error={error} onRetry={() => void refresh()}>
        <div className="space-y-6">
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
        </div>
      </DataSection>
    </PageShell>
  )
}

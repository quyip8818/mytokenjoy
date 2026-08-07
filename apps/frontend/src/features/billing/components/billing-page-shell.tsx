import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import type { useBillingPage } from '@/features/billing'
import { PermissionGate, useSession } from '@/features/session'
import { useWorkflow } from '@/features/workflow'
import { PERMISSION } from '@/lib/permissions'
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
  const { companyId, companyName, companyType } = useSession()
  const { open } = useWorkflow()
  const canRecharge = companyType !== 'trial' && companyType !== 'demo'

  return (
    <PageShell>
      <PageHeader testId="page-billing" title="钱包管理" description="账户余额与充值管理" />

      <BillingStats wallet={wallet} loading={loading} />

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() =>
            open('lot-audit', {
              companyId,
              companyName,
              readonly: true,
            })
          }
        >
          查看批次明细
        </Button>
      </div>

      <DataSection loading={loading} error={error} onRetry={() => void refresh()}>
        <div className="space-y-6">
          {canRecharge ? (
            <PermissionGate
              permission={PERMISSION.BILLING_MANAGE}
              fallback={
                <div className="rounded-lg border border-border bg-card p-5">
                  <p className="text-sm text-muted-foreground">
                    当前角色无充值权限，如需充值请联系管理员。
                  </p>
                </div>
              }
            >
              <RechargePanel
                currency={walletBillingCurrency(wallet)}
                rechargePending={rechargePending}
                onRecharge={handleRecharge}
              />
            </PermissionGate>
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

import { useState } from 'react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import type { useBillingPage } from '@/features/billing'
import { PermissionGate, useSession } from '@/features/session'
import { useWorkflow } from '@/features/workflow'
import { PERMISSION } from '@/lib/permissions'
import { IS_SAAS } from '@/config/app'
import { walletBillingCurrency } from '../lib/selectors'
import { RechargeRecordsTable } from './recharge-records-table'
import { BillingStats } from './billing-stats'

type BillingPageShellProps = ReturnType<typeof useBillingPage>

export function BillingPageShell({
  wallet,
  loading,
  error,
  refresh,
  topUpRecords,
}: BillingPageShellProps) {
  const { companyId, companyName, companyType } = useSession()
  const { open } = useWorkflow()
  const [showLocalHint, setShowLocalHint] = useState(false)

  // ponytail: SaaS 上 platform/demo 账户不能充值；local 版弹提示
  const canRecharge = IS_SAAS && companyType !== 'platform' && companyType !== 'demo'

  const handleRechargeClick = () => {
    if (IS_SAAS) {
      open('recharge', {
        currency: walletBillingCurrency(wallet),
        onSuccess: () => void refresh(),
      })
    } else {
      setShowLocalHint(true)
    }
  }

  return (
    <PageShell>
      <PageHeader
        testId="page-billing"
        title="钱包管理"
        description="账户余额与充值管理"
        actions={
          canRecharge ? (
            <PermissionGate write permission={PERMISSION.BILLING_MANAGE}>
              <Button variant="brand" onClick={handleRechargeClick}>
                充值
              </Button>
            </PermissionGate>
          ) : !IS_SAAS ? (
            <Button variant="brand" onClick={handleRechargeClick}>
              充值
            </Button>
          ) : undefined
        }
      />

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
        <RechargeRecordsTable records={topUpRecords} />
      </DataSection>

      {/* ponytail: selfhosted 无在线支付，引导去 SaaS 平台充值 */}
      <AlertDialog open={showLocalHint} onOpenChange={setShowLocalHint}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>在线充值</AlertDialogTitle>
            <AlertDialogDescription>
              请使用公司管理员账户登录{' '}
              <a
                href="https://www.tokenjoy.me/billing"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-primary underline underline-offset-4"
              >
                www.tokenjoy.me/billing
              </a>{' '}
              进行充值。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={() => setShowLocalHint(false)}>知道了</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageShell>
  )
}

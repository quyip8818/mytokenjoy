import { PermissionGate } from '@/components/auth/permission-gate'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PERMISSION } from '@/lib/permission-keys'
import { useBillingPage } from '@/routes/billing/hooks/use-billing-page'

export default function BillingPage() {
  const { wallet, loading, canWrite, rechargePending, handleRecharge } = useBillingPage()

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-xl font-semibold">企业钱包</h1>
      <Card>
        <CardHeader>
          <CardTitle>余额</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : (
            <p className="text-2xl font-medium">
              {wallet?.availableQuota?.toLocaleString() ?? '—'}{' '}
              <span className="text-sm font-normal text-muted-foreground">
                {wallet?.currency ?? 'CNY'}
              </span>
            </p>
          )}
          <PermissionGate permission={PERMISSION.BILLING_RECHARGE} write>
            <Button type="button" onClick={handleRecharge} disabled={!canWrite || rechargePending}>
              充值 100
            </Button>
          </PermissionGate>
        </CardContent>
      </Card>
    </div>
  )
}

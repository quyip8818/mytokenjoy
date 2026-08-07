import { useState } from 'react'
import { toast } from '@/lib/toast'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { LotAuditEntry } from '@/api/types'

export function useLotAudit(companyId: string, readonly: boolean) {
  const apis = useInjectedApis()

  const { data, loading, refresh } = useInjectedQuery({
    queryKey: ['lot-audit', companyId],
    queryFn: (a) =>
      readonly
        ? a.billingApi.listLots().then((lots) => ({ lots, walletRemainQuota: 0 }))
        : a.platformApi.listCompanyLots(companyId),
  })

  const lots: LotAuditEntry[] = data?.lots ?? []
  const walletRemainQuota: number = data?.walletRemainQuota ?? 0

  // Refund state
  const [refundTarget, setRefundTarget] = useState<LotAuditEntry | null>(null)
  const [refundAmount, setRefundAmount] = useState('')
  const [refunding, setRefunding] = useState(false)

  const openRefund = (lot: LotAuditEntry) => {
    setRefundTarget(lot)
    const maxMoney = lot.quotaRemaining / lot.quotaPerUnit
    setRefundAmount(String(Math.floor(maxMoney * 100) / 100))
  }

  const closeRefund = () => setRefundTarget(null)

  const handleRefund = async () => {
    if (!refundTarget) return
    const amount = Number(refundAmount)
    if (amount <= 0) {
      toast.error('请输入有效金额')
      return
    }
    const maxMoney = refundTarget.quotaRemaining / refundTarget.quotaPerUnit
    if (amount > maxMoney + 0.01) {
      toast.error(`退费金额不能超过剩余 ¥${maxMoney.toFixed(2)}`)
      return
    }
    setRefunding(true)
    try {
      await apis.platformApi.refundCompany(companyId, refundTarget.id, amount)
      toast.success(`已退费 ¥${amount}`)
      setRefundTarget(null)
      void refresh()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : '退费失败')
    } finally {
      setRefunding(false)
    }
  }

  return {
    lots,
    walletRemainQuota,
    loading,
    refresh,
    refundTarget,
    refundAmount,
    setRefundAmount,
    refunding,
    openRefund,
    closeRefund,
    handleRefund,
  }
}

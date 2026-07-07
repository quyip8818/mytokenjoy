import { useState, useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import { toast } from 'sonner'
import { toTopUpRecordView } from '../lib/mappers'

export type PaymentMethod = 'alipay' | 'wechat'

export interface TopUpRecordView {
  id: string
  orderId: string
  method: PaymentMethod
  amount: number
  paidAmount: number
  invoiceStatus: 'none' | 'applied' | 'issued'
  createdAt: string
}

export function useWalletPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [rechargePending, setRechargePending] = useState(false)

  const {
    data: wallet,
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.billing.wallet(),
    queryFn: (api) => api.billingApi.getWallet(),
  })

  const { data: records, refresh: refreshRecords } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.billing.rechargeRecords(),
    queryFn: (api) => api.billingApi.listRechargeRecords(),
  })

  const topUpRecords = (records ?? []).map(toTopUpRecordView)

  const handleRecharge = useCallback(
    async (amount: number) => {
      if (amount <= 0) {
        toast.error('请输入有效充值金额')
        return
      }
      setRechargePending(true)
      try {
        const order = await apis.billingApi.recharge({
          amount,
          idempotencyKey: crypto.randomUUID(),
        })
        await apis.billingApi.confirmRecharge(order.id)
        toast.success('充值成功')
        await Promise.all([refresh(), refreshRecords()])
      } catch {
        toast.error('充值失败，请重试')
      } finally {
        setRechargePending(false)
      }
    },
    [apis, refresh, refreshRecords],
  )

  return {
    wallet,
    loading,
    error,
    refresh,
    topUpRecords,
    rechargePending,
    handleRecharge,
    balance: wallet?.balance ?? 0,
    currency: wallet?.currency ?? 'CNY',
    totalConsumed: wallet?.totalConsumed ?? 0,
    totalRequests: wallet?.totalRequests ?? 0,
  }
}

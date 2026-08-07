import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
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

export function useBillingPage(injectedApis?: AppApis) {
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

  const { data: records } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.billing.rechargeRecords(),
    queryFn: (api) => api.billingApi.listRechargeRecords(),
  })

  const topUpRecords = (records ?? []).map(toTopUpRecordView)

  return {
    wallet,
    loading,
    error,
    refresh,
    topUpRecords,
  }
}

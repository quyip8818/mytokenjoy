import { request } from './client'
import type { RechargeInput, RechargeOrder, TopUpRecord, WalletView } from './types'

export const billingApi = {
  getWallet: () => request<WalletView>('/billing/wallet'),
  listRechargeRecords: () => request<TopUpRecord[]>('/billing/recharge-records'),
  recharge: (input: RechargeInput) =>
    request<RechargeOrder>('/billing/recharge', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  confirmRecharge: (orderId: string) =>
    request<void>(`/billing/recharge/${orderId}/confirm`, { method: 'POST' }),
}

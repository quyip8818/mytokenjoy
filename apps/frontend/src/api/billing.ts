import { request } from './client'
import type {
  DiscountEntry,
  LotAuditEntry,
  RechargeInput,
  RechargeOrder,
  TopUpRecord,
  WalletView,
} from './types'

export interface RedeemResult {
  faceValue: number
  currency: string
  quotaGranted: number
}

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
  redeem: (code: string) =>
    request<RedeemResult>('/billing/redeem', {
      method: 'POST',
      body: JSON.stringify({ code }),
    }),
  myDiscounts: () => request<DiscountEntry[]>('/billing/discounts'),
  listLots: () => request<LotAuditEntry[]>('/billing/lots'),
}

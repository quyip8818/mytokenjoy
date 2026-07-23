import { request } from './client'

export interface WalletCurrencyView {
  currency: string
  balance: number
  totalTopup: number
  totalConsumed: number
}

export interface WalletView {
  companyId: string
  billingCurrency: string
  balances: WalletCurrencyView[]
  walletQuotaRemain: number
  giftQuota: number
  overdraftQuota: number
  totalRequests: number
}

export interface TopUpRecord {
  id: string
  orderId: string
  method: 'alipay' | 'wechat'
  amount: number
  paidAmount: number
  invoiceStatus: 'none' | 'applied' | 'issued'
  status: 'pending' | 'confirmed' | 'failed'
  createdAt: string
}

export interface RechargeInput {
  amount: number
  idempotencyKey: string
}

export interface RechargeOrder {
  id: string
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
}

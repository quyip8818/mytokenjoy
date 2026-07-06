import { request } from './client'

export interface WalletView {
  companyId: number
  currency: string
  balance: number
  allocatable?: number
  totalConsumed: number
  totalRequests: number
}

export interface TopUpRecord {
  id: string
  orderId: string
  method: 'alipay' | 'wechat'
  amount: number
  paidAmount: number
  invoiceStatus: 'none' | 'applied' | 'issued'
  status: 'success' | 'pending' | 'failed'
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

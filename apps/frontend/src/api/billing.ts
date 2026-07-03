import { request } from './client'

export interface WalletView {
  companyId: number
  currency: string
  availableQuota: number
}

export interface RechargeInput {
  amount: number
  idempotencyKey: string
}

export const billingApi = {
  getWallet: () => request<WalletView>('/billing/wallet'),
  recharge: (input: RechargeInput) =>
    request<{ orderId: string }>('/billing/recharge', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
}

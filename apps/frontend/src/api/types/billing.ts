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
  walletRemainQuota: number
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

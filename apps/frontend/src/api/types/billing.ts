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

export interface DiscountEntry {
  modelType: string
  discount: number
  note?: string
}

export interface LotTransaction {
  id: string
  action: 'credit' | 'refund'
  quotaDelta: number
  moneyAmount: number
  remainingAfter: number
  source: string
  operatorId?: string
  note: string
  createdAt: string
}

export interface LotAuditEntry {
  id: string
  orderId: string
  lotKind: 'paid' | 'gift' | 'adjust' | 'overdraft' | 'mock'
  billingCurrency: string
  quotaPerUnit: number
  quotaGranted: number
  quotaRemaining: number
  paidAmount: number
  status: 'active' | 'exhausted'
  createdAt: string
  transactions: LotTransaction[]
}

export interface LotAuditResponse {
  lots: LotAuditEntry[]
  walletRemainQuota: number
}

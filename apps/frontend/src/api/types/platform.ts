export interface PlatformCompanyWallet {
  balance: number
  giftBalance: number
  overdraft: number
  totalTopup: number
  totalConsumed: number
}

export interface PlatformCompanyOverview {
  id: string
  name: string
  type: string
  status: string
  billingCurrency: string
  wallet: PlatformCompanyWallet
  monthlySpend: number
  memberCount: number
  createdAt: string
}

export interface PlatformModel {
  modelId: string
  provider: string
  type: string
  name: string
  description: string
  inputPrice: number
  outputPrice: number
  maxContext: number
  active: boolean
  capabilities: string[]
  source: string
}

export interface PlatformCreateModelInput {
  type: string
  name: string
  provider: string
  inputPrice: number
  outputPrice: number
  capabilities?: string[]
  maxContext?: number
}

export interface PlatformUpdateModelInput {
  name?: string
  type?: string
  provider?: string
  active?: boolean
  capabilities?: string[]
  maxContext?: number
}

export interface PlatformSetPricingInput {
  inputPrice: number
  outputPrice: number
}

export interface PlatformCurrency {
  code: string
  quotaPerUnit: number
  enabled: boolean
  updatedAt: string
}

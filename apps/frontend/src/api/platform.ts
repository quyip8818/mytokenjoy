import { request } from './client'

// --- Types ---

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

// --- API ---

export const platformApi = {
  // --- Models ---
  listModels: () => request<PlatformModel[]>('/platform/models'),

  createModel: (data: PlatformCreateModelInput) =>
    request<PlatformModel>('/platform/models', { method: 'POST', body: JSON.stringify(data) }),

  updateModel: (id: string, data: PlatformUpdateModelInput) =>
    request<PlatformModel>(`/platform/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  deleteModel: (id: string) => request<void>(`/platform/models/${id}`, { method: 'DELETE' }),

  setPricing: (id: string, data: PlatformSetPricingInput) =>
    request<void>(`/platform/models/${id}/pricing`, { method: 'PUT', body: JSON.stringify(data) }),

  publish: () => request<{ version: number }>('/platform/catalog/publish', { method: 'POST' }),

  // --- Companies ---
  companiesOverview: () => request<PlatformCompanyOverview[]>('/platform/companies/overview'),

  rechargeCompany: (id: string, amount: number) =>
    request<void>(`/platform/companies/${id}/recharge`, {
      method: 'POST',
      body: JSON.stringify({ amount }),
    }),

  giftCompany: (id: string, amount: number) =>
    request<void>(`/platform/companies/${id}/gift`, {
      method: 'POST',
      body: JSON.stringify({ amount }),
    }),

  updateCompany: (id: string, patch: { status?: string }) =>
    request<void>(`/platform/companies/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),

  // --- Currencies ---
  listCurrencies: () => request<PlatformCurrency[]>('/platform/currencies'),

  createCurrency: (data: { code: string; quotaPerUnit: number }) =>
    request<PlatformCurrency>('/platform/currencies', { method: 'POST', body: JSON.stringify(data) }),

  updateCurrency: (code: string, data: { quotaPerUnit: number }) =>
    request<PlatformCurrency>(`/platform/currencies/${code}`, { method: 'PUT', body: JSON.stringify(data) }),

  toggleCurrencyStatus: (code: string, enabled: boolean) =>
    request<PlatformCurrency>(`/platform/currencies/${code}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ enabled }),
    }),
}

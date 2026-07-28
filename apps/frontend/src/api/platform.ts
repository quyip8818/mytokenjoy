import { request } from './client'

// --- Types ---

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

// --- API ---

export const platformApi = {
  listModels: () => request<PlatformModel[]>('/platform/models'),

  createModel: (data: PlatformCreateModelInput) =>
    request<PlatformModel>('/platform/models', { method: 'POST', body: JSON.stringify(data) }),

  updateModel: (id: string, data: PlatformUpdateModelInput) =>
    request<PlatformModel>(`/platform/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  deleteModel: (id: string) =>
    request<void>(`/platform/models/${id}`, { method: 'DELETE' }),

  setPricing: (id: string, data: PlatformSetPricingInput) =>
    request<void>(`/platform/models/${id}/pricing`, { method: 'PUT', body: JSON.stringify(data) }),

  publish: () => request<{ version: number }>('/platform/catalog/publish', { method: 'POST' }),
}

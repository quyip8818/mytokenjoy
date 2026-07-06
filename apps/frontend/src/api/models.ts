import { request } from './client'
import type { ModelInfo, RoutingRule } from './types'

export const modelApi = {
  list: () => request<ModelInfo[]>('/models'),
  create: (data: Omit<ModelInfo, 'id'>) =>
    request<ModelInfo>('/models', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<ModelInfo>) =>
    request<ModelInfo>(`/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  toggle: (id: string, enabled: boolean) =>
    request<void>(`/models/${id}/toggle`, { method: 'PUT', body: JSON.stringify({ enabled }) }),
  delete: (id: string) =>
    request<void>(`/models/${id}`, { method: 'DELETE' }),
}

export const routingApi = {
  getRules: () => request<RoutingRule[]>('/models/routing'),
  updateRule: (id: string, data: { allowedModels: string[]; defaultModel: string | null; fallbackModel: string | null }) =>
    request<RoutingRule>(`/models/routing/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
}

import { request } from './client'
import type { CreateModelInput, ModelInfo, ResolvedWhitelist, RoutingRule } from './types'

export const modelApi = {
  list: () => request<ModelInfo[]>('/models'),
  create: (data: CreateModelInput) =>
    request<ModelInfo>('/models', { method: 'POST', body: JSON.stringify(data) }),
  toggle: (id: string, enabled: boolean) =>
    request<void>(`/models/${id}/toggle`, { method: 'PUT', body: JSON.stringify({ enabled }) }),
}

export const routingApi = {
  getRules: () => request<RoutingRule[]>('/models/routing'),
  updateRule: (
    id: string,
    data: {
      allowedModels: string[]
      inherited: boolean
      defaultModel?: string | null
      fallbackModel?: string | null
    },
  ) => request<RoutingRule>(`/models/routing/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  resolveWhitelist: (deptId: string) =>
    request<ResolvedWhitelist>(`/models/routing/resolve?deptId=${encodeURIComponent(deptId)}`),
}

import { request } from './client'
import type {
  CreateModelInput,
  ModelInfo,
  ResolvedWhitelist,
  RoutingRule,
  UpdateModelInput,
} from './types'

export const modelApi = {
  list: () => request<ModelInfo[]>('/models'),
  create: (data: CreateModelInput) =>
    request<ModelInfo>('/models', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: UpdateModelInput) =>
    request<ModelInfo>(`/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/models/${id}`, { method: 'DELETE' }),
  toggle: (id: number, enabled: boolean) =>
    request<void>(`/models/${id}/toggle`, { method: 'PUT', body: JSON.stringify({ enabled }) }),
}

export const routingApi = {
  getRules: () => request<RoutingRule[]>('/models/routing'),
  updateRule: (
    id: string,
    data: {
      allowedModelIds: number[]
      inherited: boolean
      defaultModelId?: number | null
      fallbackModelId?: number | null
    },
  ) => request<RoutingRule>(`/models/routing/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  resolveWhitelist: (departmentId: string) =>
    request<ResolvedWhitelist>(
      `/models/routing/resolve?deptId=${encodeURIComponent(departmentId)}`,
    ),
}

import { request } from './client'
import type {
  CreateModelInput,
  ModelInfo,
  ResolvedWhitelist,
  RoutingRule,
  UpdateModelInput,
} from './types'

export const modelsApi = {
  list: () => request<ModelInfo[]>('/models'),
  create: (data: CreateModelInput) =>
    request<ModelInfo>('/models', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: UpdateModelInput) =>
    request<ModelInfo>(`/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  toggle: (id: string, enabled: boolean) =>
    request<void>(`/models/${id}/toggle`, { method: 'PUT', body: JSON.stringify({ enabled }) }),

  routing: {
    getRules: () => request<RoutingRule[]>('/models/routing'),
    updateRule: (
      id: string,
      data: {
        allowedModelIds: string[]
        inherited: boolean
        defaultModelId?: string | null
        fallbackModelId?: string | null
      },
    ) =>
      request<RoutingRule>(`/models/routing/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    resolveWhitelist: (departmentId: string) =>
      request<ResolvedWhitelist>(
        `/models/routing/resolve?deptId=${encodeURIComponent(departmentId)}`,
      ),
  },
}

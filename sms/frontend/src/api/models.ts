import { request, buildQuery } from './client'

export interface AiModel {
  id: string
  supplierId: string
  modelName: string
  modelId?: string
  modelType?: string
  contextLength?: number
  inputPrice?: number
  outputPrice?: number
  discount?: number
  status: string
  description?: string
  createdAt: string
  updatedAt: string
  supplierName?: string
}

export interface ModelCreateInput {
  supplierId: string
  modelName: string
  modelId?: string
  modelType?: string
  contextLength?: number
  inputPrice?: number
  outputPrice?: number
  discount?: number
  status: string
  description?: string
}

export interface ModelUpdateInput {
  supplierId?: string
  modelName?: string
  modelType?: string
  contextLength?: number
  inputPrice?: number
  outputPrice?: number
  discount?: string
  status?: string
  description?: string
}

export const modelsApi = {
  list: (params: Record<string, unknown>) =>
    request<{ items: AiModel[]; total: number; page: number; pageSize: number }>(
      `/models${buildQuery(params)}`,
    ),
  detail: (id: string) => request<AiModel>(`/models/${id}`),
  create: (data: ModelCreateInput) =>
    request<AiModel>('/models', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: ModelUpdateInput) =>
    request<AiModel>(`/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/models/${id}`, { method: 'DELETE' }),
}

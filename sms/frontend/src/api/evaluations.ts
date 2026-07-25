import { request, buildQuery } from './client'

export interface Evaluation {
  id: string
  supplierId: string
  evaluatorId: string
  period: string
  quality: number
  performance: number
  price: number
  service: number
  compliance: number
  totalScore: number
  grade: string
  comment?: string
  createdAt: string
  supplierName?: string
  evaluatorName?: string
}

export interface EvaluationWeight {
  id: string
  dimension: string
  weight: number
}

export interface EvaluationCreateInput {
  supplierId: string
  period: string
  quality: number
  performance: number
  price: number
  service: number
  compliance: number
  comment?: string
}

export interface EvaluationUpdateInput {
  quality?: number
  performance?: number
  price?: number
  service?: number
  compliance?: number
  comment?: string
}

export const evaluationsApi = {
  list: (params: Record<string, unknown>) =>
    request<{ items: Evaluation[]; total: number; page: number; pageSize: number }>(
      `/evaluations${buildQuery(params)}`,
    ),
  detail: (id: string) => request<Evaluation>(`/evaluations/${id}`),
  create: (data: EvaluationCreateInput) =>
    request<Evaluation>('/evaluations', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: EvaluationUpdateInput) =>
    request<Evaluation>(`/evaluations/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/evaluations/${id}`, { method: 'DELETE' }),
  getWeights: () => request<EvaluationWeight[]>('/evaluations/weights'),
  updateWeights: (weights: EvaluationWeight[]) =>
    request<void>('/evaluations/weights', { method: 'PUT', body: JSON.stringify(weights) }),
}

import { request } from './client'
import type {
  CostSummary,
  DailyCost,
  DepartmentCost,
  ModelUsage,
  TeamUsage,
  TopConsumer,
} from './types'

export const dashboardApi = {
  getCostSummary: (period?: string) =>
    request<CostSummary>(`/dashboard/cost/summary${period ? `?period=${period}` : ''}`),
  getDepartmentCosts: () => request<DepartmentCost[]>('/dashboard/cost/departments'),
  getDailyCosts: (days?: number) =>
    request<DailyCost[]>(`/dashboard/cost/daily${days ? `?days=${days}` : ''}`),
  getTopConsumers: (limit?: number) =>
    request<TopConsumer[]>(`/dashboard/cost/top${limit ? `?limit=${limit}` : ''}`),
  getModelUsage: () => request<ModelUsage[]>('/dashboard/usage/models'),
  getTeamUsage: () => request<TeamUsage[]>('/dashboard/usage/teams'),
}

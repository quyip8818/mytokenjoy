import { request, buildQuery } from './client'
import type {
  CostQueryParams,
  CostSummary,
  DailyCost,
  DepartmentCost,
  DepartmentCostMember,
  ModelUsage,
  TeamUsage,
  TopConsumer,
} from './types'

export const dashboardApi = {
  getCostSummary: (params?: CostQueryParams) =>
    request<CostSummary>(`/dashboard/cost/summary${buildQuery(params ?? {})}`),
  getDepartmentCosts: (params?: CostQueryParams & { parentId?: string }) =>
    request<DepartmentCost[]>(`/dashboard/cost/departments${buildQuery(params ?? {})}`),
  getDepartmentMemberCosts: (deptId: string, params?: CostQueryParams) =>
    request<DepartmentCostMember[]>(
      `/dashboard/cost/departments/${deptId}/members${buildQuery(params ?? {})}`,
    ),
  getDailyCosts: (params?: CostQueryParams) =>
    request<DailyCost[]>(`/dashboard/cost/daily${buildQuery(params ?? {})}`),
  getTopConsumers: (params?: CostQueryParams & { limit?: number }) =>
    request<TopConsumer[]>(`/dashboard/cost/top${buildQuery(params ?? {})}`),
  getModelUsage: () => request<ModelUsage[]>('/dashboard/usage/models'),
  getTeamUsage: () => request<TeamUsage[]>('/dashboard/usage/teams'),
}

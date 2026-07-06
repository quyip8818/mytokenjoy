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
  UsageSeriesQuery,
  UsageSeriesResponse,
} from './types'

export const dashboardApi = {
  getCostSummary: (params?: CostQueryParams) =>
    request<CostSummary>(`/dashboard/cost/summary${buildQuery(params ?? {})}`),
  getDepartmentCosts: (params?: CostQueryParams & { parentId?: string }) =>
    request<DepartmentCost[]>(`/dashboard/cost/departments${buildQuery(params ?? {})}`),
  getDepartmentMemberCosts: (departmentId: string, params?: CostQueryParams) =>
    request<DepartmentCostMember[]>(
      `/dashboard/cost/departments/${departmentId}/members${buildQuery(params ?? {})}`,
    ),
  getDailyCosts: (params?: CostQueryParams) =>
    request<DailyCost[]>(`/dashboard/cost/daily${buildQuery(params ?? {})}`),
  getTopConsumers: (params?: CostQueryParams & { limit?: number }) =>
    request<TopConsumer[]>(`/dashboard/cost/top${buildQuery(params ?? {})}`),
  getModelUsage: (params?: CostQueryParams) =>
    request<ModelUsage[]>(`/dashboard/usage/models${buildQuery(params ?? {})}`),
  getTeamUsage: (params?: CostQueryParams) =>
    request<TeamUsage[]>(`/dashboard/usage/teams${buildQuery(params ?? {})}`),
  getUsageSeries: (params: UsageSeriesQuery) =>
    request<UsageSeriesResponse>(`/dashboard/usage/series${buildQuery(params)}`),
}

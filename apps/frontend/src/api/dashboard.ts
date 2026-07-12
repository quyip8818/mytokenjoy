import { request, buildQuery } from './client'
import type {
  CostQueryParams,
  CostSummary,
  DailyCost,
  DepartmentCost,
  DepartmentCostMember,
  ModelUsage,
  DepartmentUsage,
  TopConsumer,
} from './types'

export const dashboardApi = {
  getCostSummary: (params?: CostQueryParams & { departmentId?: string }) =>
    request<CostSummary>(`/dashboard/cost/summary${buildQuery(params ?? {})}`),
  getDepartmentCosts: (params?: CostQueryParams & { parentId?: string }) =>
    request<DepartmentCost[]>(`/dashboard/cost/departments${buildQuery(params ?? {})}`),
  getDepartmentMemberCosts: (departmentId: string, params?: CostQueryParams) =>
    request<DepartmentCostMember[]>(
      `/dashboard/cost/departments/${departmentId}/members${buildQuery(params ?? {})}`,
    ),
  getDailyCosts: (params?: CostQueryParams & { departmentId?: string }) =>
    request<DailyCost[]>(`/dashboard/cost/daily${buildQuery(params ?? {})}`),
  getTopConsumers: (params?: CostQueryParams & { limit?: number; departmentId?: string }) =>
    request<TopConsumer[]>(`/dashboard/cost/top${buildQuery(params ?? {})}`),
  getModelUsage: (params?: CostQueryParams & { departmentId?: string }) =>
    request<ModelUsage[]>(`/dashboard/usage/models${buildQuery(params ?? {})}`),
  getDepartmentUsage: (params?: CostQueryParams & { departmentId?: string }) =>
    request<DepartmentUsage[]>(`/dashboard/usage/teams${buildQuery(params ?? {})}`),
}

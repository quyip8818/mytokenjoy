import { request, buildQuery } from './client'
import type {
  CostPeriod,
  CostSummary,
  DailyCost,
  DepartmentCost,
  DepartmentCostMember,
  ModelUsage,
  TeamUsage,
  TopConsumer,
} from './types'

export const dashboardApi = {
  getCostSummary: (period?: CostPeriod) =>
    request<CostSummary>(`/dashboard/cost/summary${buildQuery({ period })}`),
  getDepartmentCosts: (params?: { parentId?: string; period?: CostPeriod }) =>
    request<DepartmentCost[]>(`/dashboard/cost/departments${buildQuery(params ?? {})}`),
  getDepartmentMemberCosts: (deptId: string, period?: CostPeriod) =>
    request<DepartmentCostMember[]>(
      `/dashboard/cost/departments/${deptId}/members${buildQuery({ period })}`,
    ),
  getDailyCosts: (period?: CostPeriod) =>
    request<DailyCost[]>(`/dashboard/cost/daily${buildQuery({ period })}`),
  getTopConsumers: (params?: { limit?: number; period?: CostPeriod }) =>
    request<TopConsumer[]>(`/dashboard/cost/top${buildQuery(params ?? {})}`),
  getModelUsage: () => request<ModelUsage[]>('/dashboard/usage/models'),
  getTeamUsage: () => request<TeamUsage[]>('/dashboard/usage/teams'),
}

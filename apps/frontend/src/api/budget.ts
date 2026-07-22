import { request } from './client'
import type {
  AlertRule,
  Project,
  BudgetNode,
  MemberBudget,
  OverrunPolicyConfig,
  UpdateMemberBudgetInput,
} from './types'

function updateDepartmentRequest(
  departmentId: string,
  data: { budget: number; reservedPool?: number },
) {
  return request<BudgetNode>(`/budget/departments/${departmentId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export const budgetApi = {
  getTree: (period?: string) =>
    request<BudgetNode[]>(`/budget/tree${period ? `?period=${period}` : ''}`),
  updateDepartment: updateDepartmentRequest,
  getMemberBudgets: (departmentId: string) =>
    request<MemberBudget[]>(`/budget/departments/${departmentId}/member-budgets`),
  updateMemberBudget: (memberId: string, data: UpdateMemberBudgetInput) =>
    request<MemberBudget>(`/budget/members/${memberId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  applyAverageBudget: (
    departmentId: string,
    data: { personalBudget: number; recursive: boolean },
  ) =>
    request<void>(`/budget/departments/${departmentId}/apply-average-budget`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  getProjects: () => request<Project[]>('/budget/projects'),
  createProject: (data: Omit<Project, 'id' | 'consumed'>) =>
    request<Project>('/budget/projects', { method: 'POST', body: JSON.stringify(data) }),
  updateProject: (id: string, data: Partial<Omit<Project, 'id' | 'consumed'>>) =>
    request<Project>(`/budget/projects/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteProject: (id: string) => request<void>(`/budget/projects/${id}`, { method: 'DELETE' }),
  getOverrunPolicy: () => request<OverrunPolicyConfig>('/budget/overrun-policy'),
  updateOverrunPolicy: (data: OverrunPolicyConfig) =>
    request<OverrunPolicyConfig>('/budget/overrun-policy', {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  getAlerts: () => request<AlertRule[]>('/budget/alerts'),
  updateAlert: (id: string, data: Partial<AlertRule>) =>
    request<AlertRule>(`/budget/alerts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  createAlert: (data: Omit<AlertRule, 'id'>) =>
    request<AlertRule>('/budget/alerts', { method: 'POST', body: JSON.stringify(data) }),
  deleteAlert: (id: string) => request<void>(`/budget/alerts/${id}`, { method: 'DELETE' }),
  getProjectMemberConsumed: (projectId: string) =>
    request<Record<string, number>>(`/budget/projects/${projectId}/member-consumed`),
}

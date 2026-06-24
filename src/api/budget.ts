import { request } from './client'
import type { AlertRule, BudgetGroup, BudgetNode, OverrunPolicyConfig } from './types'

export const budgetApi = {
  getTree: (period?: string) =>
    request<BudgetNode[]>(`/budget/tree${period ? `?period=${period}` : ''}`),
  updateNode: (id: string, data: { budget: number; reservedPool?: number }) =>
    request<BudgetNode>(`/budget/nodes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  getGroups: () => request<BudgetGroup[]>('/budget/groups'),
  createGroup: (data: Omit<BudgetGroup, 'id' | 'consumed'>) =>
    request<BudgetGroup>('/budget/groups', { method: 'POST', body: JSON.stringify(data) }),
  updateGroup: (id: string, data: Partial<Omit<BudgetGroup, 'id' | 'consumed'>>) =>
    request<BudgetGroup>(`/budget/groups/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteGroup: (id: string) => request<void>(`/budget/groups/${id}`, { method: 'DELETE' }),
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
}

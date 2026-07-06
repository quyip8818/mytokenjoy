import { request } from './client'
import type { AlertRule, BudgetApproval, BudgetNode, BudgetProject } from './types'

export const budgetApi = {
  getTree: (period?: string) =>
    request<BudgetNode[]>(`/budget/tree${period ? `?period=${period}` : ''}`),
  updateNode: (id: string, data: Partial<Pick<BudgetNode, 'budget' | 'reserved' | 'memberQuota' | 'overrunPolicy'>>) =>
    request<BudgetNode>(`/budget/nodes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  getProjects: (period?: string) =>
    request<BudgetProject[]>(`/budget/projects${period ? `?period=${period}` : ''}`),
  createProject: (data: Omit<BudgetProject, 'id' | 'consumed'>) =>
    request<BudgetProject>('/budget/projects', { method: 'POST', body: JSON.stringify(data) }),
  updateProject: (id: string, data: Partial<BudgetProject>) =>
    request<BudgetProject>(`/budget/projects/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteProject: (id: string) =>
    request<void>(`/budget/projects/${id}`, { method: 'DELETE' }),

  getApprovals: () => request<BudgetApproval[]>('/budget/approvals'),
  resolveApproval: (id: string, data: { status: 'approved' | 'rejected'; rejectReason?: string }) =>
    request<BudgetApproval>(`/budget/approvals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  getAlerts: () => request<AlertRule[]>('/budget/alerts'),
  updateAlert: (id: string, data: Partial<AlertRule>) =>
    request<AlertRule>(`/budget/alerts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  createAlert: (data: Omit<AlertRule, 'id'>) =>
    request<AlertRule>('/budget/alerts', { method: 'POST', body: JSON.stringify(data) }),
  deleteAlert: (id: string) =>
    request<void>(`/budget/alerts/${id}`, { method: 'DELETE' }),
}

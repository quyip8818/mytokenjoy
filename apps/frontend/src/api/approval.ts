import { request, buildQuery } from './client'
import type {
  ApprovalListResponse,
  ApprovalPreCheck,
  ApprovalRequest,
  ApprovalStatus,
  ApprovalType,
} from './types'

export const approvalApi = {
  list: (params?: {
    status?: ApprovalStatus
    type?: ApprovalType
    applicantId?: string
    limit?: number
    offset?: number
  }) => request<ApprovalListResponse>(`/approvals${buildQuery(params ?? {})}`),

  get: (id: string) => request<ApprovalRequest>(`/approvals/${id}`),

  create: (data: { type: ApprovalType; metadata: Record<string, unknown> }) =>
    request<ApprovalRequest>('/approvals', { method: 'POST', body: JSON.stringify(data) }),

  approve: (id: string) => request<void>(`/approvals/${id}/approve`, { method: 'PUT' }),

  reject: (id: string, reason: string) =>
    request<void>(`/approvals/${id}/reject`, { method: 'PUT', body: JSON.stringify({ reason }) }),

  cancel: (id: string) => request<void>(`/approvals/${id}/cancel`, { method: 'PUT' }),

  retry: (id: string) => request<void>(`/approvals/${id}/retry`, { method: 'PUT' }),

  preCheck: (id: string) => request<ApprovalPreCheck>(`/approvals/${id}/pre-check`),
}

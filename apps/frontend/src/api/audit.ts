import { request, buildQuery } from './client'
import type { AuditSettings, CallLog, OperationLog, Paginated } from './types'

export const auditApi = {
  getOperations: (params?: { page?: number; pageSize?: number; action?: string }) =>
    request<Paginated<OperationLog>>(`/audit/operations${buildQuery(params ?? {})}`),
  getCalls: (params?: { page?: number; pageSize?: number; model?: string; status?: string }) =>
    request<Paginated<CallLog>>(`/audit/calls${buildQuery(params ?? {})}`),
  getSettings: () => request<AuditSettings>('/audit/settings'),
  updateSettings: (data: AuditSettings) =>
    request<AuditSettings>('/audit/settings', { method: 'PUT', body: JSON.stringify(data) }),
}

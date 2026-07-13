import { request, buildQuery } from './client'
import type {
  AuditCallsQueryParams,
  AuditOperationsQueryParams,
  AuditSettings,
  CallLog,
  CallsSummary,
  OperationDailyCount,
  OperationLog,
  Paginated,
} from './types'

export const auditApi = {
  getOperations: (params?: AuditOperationsQueryParams) =>
    request<Paginated<OperationLog>>(`/audit/operations${buildQuery(params ?? {})}`),
  getOperationsTimeline: (params?: AuditOperationsQueryParams) =>
    request<OperationDailyCount[]>(`/audit/operations/timeline${buildQuery(params ?? {})}`),
  getCalls: (params?: AuditCallsQueryParams) =>
    request<Paginated<CallLog>>(`/audit/calls${buildQuery(params ?? {})}`),
  getCallsSummary: (params?: AuditCallsQueryParams) =>
    request<CallsSummary>(`/audit/calls/summary${buildQuery(params ?? {})}`),
  getSettings: () => request<AuditSettings>('/audit/settings'),
  updateSettings: (data: AuditSettings) =>
    request<AuditSettings>('/audit/settings', { method: 'PUT', body: JSON.stringify(data) }),
}

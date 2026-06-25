import { request, buildQuery } from './client'
import type {
  AuditCallsQueryParams,
  AuditOperationsQueryParams,
  AuditSettings,
  CallLog,
  OperationLog,
  Paginated,
} from './types'

export const auditApi = {
  getOperations: (params?: AuditOperationsQueryParams) =>
    request<Paginated<OperationLog>>(`/audit/operations${buildQuery(params ?? {})}`),
  getCalls: (params?: AuditCallsQueryParams) =>
    request<Paginated<CallLog>>(`/audit/calls${buildQuery(params ?? {})}`),
  getSettings: () => request<AuditSettings>('/audit/settings'),
  updateSettings: (data: AuditSettings) =>
    request<AuditSettings>('/audit/settings', { method: 'PUT', body: JSON.stringify(data) }),
}

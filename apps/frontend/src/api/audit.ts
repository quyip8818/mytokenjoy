import { request } from './client'
import type { CallLog, OperationLog, Paginated } from './types'

export const auditApi = {
  getOperations: (params?: { page?: number; pageSize?: number; action?: string; timeRange?: string }) =>
    request<Paginated<OperationLog>>(`/audit/operations?${new URLSearchParams(params as Record<string, string>)}`),
  getCalls: (params?: { page?: number; pageSize?: number; model?: string; status?: string }) =>
    request<Paginated<CallLog>>(`/audit/calls?${new URLSearchParams(params as Record<string, string>)}`),
}

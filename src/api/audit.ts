import { request, buildQuery } from './client'
import type { CallLog, OperationLog, Paginated } from './types'

export const auditApi = {
  getOperations: (params?: { page?: number; pageSize?: number; action?: string }) =>
    request<Paginated<OperationLog>>(`/audit/operations${buildQuery(params ?? {})}`),
  getCalls: (params?: { page?: number; pageSize?: number; model?: string; status?: string }) =>
    request<Paginated<CallLog>>(`/audit/calls${buildQuery(params ?? {})}`),
}

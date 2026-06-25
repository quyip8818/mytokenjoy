import type { AuditCallsQueryParams, AuditOperationsQueryParams } from '@/api/types'
import { resolveAuditDatePreset } from '@/lib/audit-constants'

export const AUDIT_FILTER_ALL = 'all'

export interface AuditBaseFilter {
  datePreset: string
  keyword: string
}

export function omitAll(value: string): string | undefined {
  return value !== AUDIT_FILTER_ALL ? value : undefined
}

export function buildAuditBaseQuery(filter: AuditBaseFilter): {
  keyword?: string
  from?: string
  to?: string
} {
  const dateRange = resolveAuditDatePreset(filter.datePreset)
  return {
    keyword: filter.keyword.trim() || undefined,
    ...dateRange,
  }
}

export interface AuditCallsFilter extends AuditBaseFilter {
  status: string
  callerId: string
}

export function buildCallsQuery(filter: AuditCallsFilter): AuditCallsQueryParams {
  return {
    ...buildAuditBaseQuery(filter),
    status: omitAll(filter.status),
    callerId: omitAll(filter.callerId),
  }
}

export interface AuditOperationsFilter extends AuditBaseFilter {
  action: string
  operatorId: string
}

export function buildOperationsQuery(filter: AuditOperationsFilter): AuditOperationsQueryParams {
  return {
    ...buildAuditBaseQuery(filter),
    action: omitAll(filter.action),
    operatorId: omitAll(filter.operatorId),
  }
}

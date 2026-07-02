import type { ProviderType } from './keys'

export type AuditAction =
  | 'key_create'
  | 'key_disable'
  | 'key_rotate'
  | 'budget_change'
  | 'budget_approve'
  | 'permission_change'
  | 'role_assign'
  | 'model_whitelist_change'
  | 'member_add'
  | 'member_remove'
  | 'org_structure_change'

export interface OperationLog {
  id: string
  action: AuditAction
  operator: string
  operatorId: string
  target: string
  detail: string
  ip: string
  createdAt: string
}

export interface CallLog {
  id: string
  caller: string
  callerId: string
  callerType: 'member' | 'platform_key'
  model: string
  provider: ProviderType
  inputTokens: number
  outputTokens: number
  latencyMs: number
  status: 'success' | 'error' | 'filtered'
  cost: number
  createdAt: string
  previewSnippet: string
}

export interface AuditSettings {
  contentRetentionEnabled: boolean
}

export interface AuditDateRangeParams {
  from?: string
  to?: string
}

export interface AuditOperationsQueryParams extends AuditDateRangeParams {
  page?: number
  pageSize?: number
  action?: string
  operatorId?: string
  keyword?: string
}

export interface AuditCallsQueryParams extends AuditDateRangeParams {
  page?: number
  pageSize?: number
  model?: string
  status?: string
  callerId?: string
  keyword?: string
}

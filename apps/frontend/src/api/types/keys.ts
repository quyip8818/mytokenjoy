export type ProviderType = 'openai' | 'anthropic' | 'deepseek' | 'qwen' | 'custom'
export type KeyStatus = 'active' | 'disabled' | 'expired' | 'error'

export interface ProviderKey {
  id: string
  provider: ProviderType
  name: string
  keyPrefix: string
  status: KeyStatus
  balance: number | null
  lastUsed: string | null
  createdAt: string
  rotateEnabled: boolean
}

export interface PlatformKey {
  id: string
  name: string
  keyPrefix: string
  fullKey?: string
  memberId: string | null
  memberName: string | null
  appName: string | null
  budgetGroupId: string | null
  budgetGroupName: string | null
  status: KeyStatus
  quota: number
  used: number
  modelWhitelist: string[]
  createdAt: string
  expiresAt: string | null
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'
export type ApprovalType = 'key' | 'quota'

export interface KeyApproval {
  id: string
  type: ApprovalType
  applicant: string
  applicantId: string
  department: string
  reason: string
  requestedQuota: number
  requestedModels: string[]
  status: ApprovalStatus
  approver: string | null
  rejectReason?: string | null
  createdAt: string
  resolvedAt: string | null
}

export interface MemberQuotaSummary {
  totalQuota: number
  used: number
  remaining: number
  reservedPool: number
}

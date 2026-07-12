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
  scope: 'member' | 'project'
  memberId: string | null
  memberName: string | null
  departmentId: string
  departmentName: string
  projectId: string | null
  projectName: string | null
  status: KeyStatus
  budget: number
  consumed: number
  modelWhitelist: number[]
  createdAt: string
  expiresAt: string | null
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'
export type ApprovalType = 'key' | 'budget'

export interface KeyApproval {
  id: string
  type: ApprovalType
  applicant: string
  applicantId: string
  department: string
  reason: string
  requestedBudget: number
  requestedModels: number[]
  status: ApprovalStatus
  approver: string | null
  rejectReason?: string | null
  createdAt: string
  resolvedAt: string | null
}

export interface MemberBudgetSummary {
  totalBudget: number
  consumed: number
  remaining: number
  reservedPool: number
}

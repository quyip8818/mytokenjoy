export type ProviderType = 'openai' | 'anthropic' | 'deepseek' | 'qwen' | 'custom'
export type KeyStatus = 'active' | 'disabled' | 'expired' | 'error'

export interface ProviderKey {
  id: string
  provider: ProviderType
  name: string
  keyPrefix: string
  status: KeyStatus
  createdAt: string
  rotateEnabled: boolean
}

export type PlatformKeyScope = 'member' | 'project' | 'project_member'

export interface PlatformKey {
  id: string
  name: string
  keyPrefix: string
  fullKey?: string
  scope: PlatformKeyScope
  memberId: string | null
  memberName: string | null
  departmentId: string
  departmentName: string
  projectId: string | null
  projectName: string | null
  status: KeyStatus
  budget: number
  consumed: number
  modelWhitelist: string[]
  createdAt: string
  expiresAt: string | null
}

export interface MemberBudgetSummary {
  totalBudget: number
  consumed: number
  remaining: number
  reservedPool: number
}

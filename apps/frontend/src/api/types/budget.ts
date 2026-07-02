export interface BudgetNode {
  id: string
  name: string
  parentId: string | null
  budget: number
  consumed: number
  reservedPool?: number
  children?: BudgetNode[]
  period: string
}

export interface OverrunPolicyConfig {
  thresholds: number[]
  notifyEmail: boolean
  notifyPhone: boolean
  notifyIm: boolean
  blockMessage: string
}

export interface ResolvedWhitelist {
  inherited: boolean
  allowedModels: string[]
  parentCount: number
}

export interface CreateModelInput {
  name: string
  displayName: string
  baseUrl: string
  apiKey: string
  inputPrice: number
  outputPrice: number
}

export interface BudgetGroup {
  id: string
  name: string
  budget: number
  consumed: number
  memberIds: string[]
  departmentIds: string[]
}

export interface AlertRule {
  id: string
  nodeId: string
  nodeName: string
  thresholds: number[]
  notifyRoleIds: string[]
  enabled: boolean
}

export interface MemberBudgetQuota {
  memberId: string
  memberName: string
  departmentId: string
  personalQuota: number
  allocated: number
  used: number
}

export interface UpdateMemberQuotaInput {
  personalQuota: number
}

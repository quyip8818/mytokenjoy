export interface BudgetNode {
  id: string
  name: string
  parentId: string | null
  budget: number
  consumed: number
  reservedPool?: number
  children?: BudgetNode[]
  period: string
  memberAvgBudget: number
}

export interface OverrunPolicyConfig {
  thresholds: number[]
  notifyEmail: boolean
  notifyPhone: boolean
  notifyIm: boolean
  blockMessage: string
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
  personalBudget: number
  allocated: number
  used: number
}

export interface UpdateMemberBudgetInput {
  personalBudget: number
}

export type OverrunPolicy = 'hard_reject' | 'approval' | 'downgrade'

export interface BudgetProjectView {
  id: string
  name: string
  budget: number
  consumed: number
  memberIds: string[]
  departmentId: string
  departmentName: string
  overrunPolicy: OverrunPolicy
  period: string
}

export interface BudgetApproval {
  id: string
  applicantName: string
  departmentName: string
  amount: number
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  createdAt: string
  resolvedAt?: string
  rejectReason?: string
}

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

export interface Project {
  id: string
  name: string
  budget: number
  consumed: number
  memberIds: string[]
  ownerDepartmentId: string
}

export interface AlertRule {
  id: string
  nodeId: string
  nodeName: string
  thresholds: number[]
  notifyRoleIds: string[]
  enabled: boolean
}

export interface MemberBudget {
  memberId: string
  memberName: string
  departmentId: string
  personalBudget: number
  allocated: number
  consumed: number
}

export interface UpdateMemberBudgetInput {
  personalBudget: number
}

export type OverrunPolicy = 'hard_reject' | 'approval' | 'downgrade'

export interface ProjectView {
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

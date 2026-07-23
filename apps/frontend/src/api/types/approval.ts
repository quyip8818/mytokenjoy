export type ApprovalStatus = 'pending' | 'approved' | 'rejected' | 'cancelled' | 'failed'
export type ApprovalType = 'key' | 'member_budget' | 'project_budget' | 'project_member_budget'

export interface ApprovalRequest {
  id: string
  type: ApprovalType
  status: ApprovalStatus
  applicantId: string
  applicantName: string
  scopeId: string
  metadata: Record<string, unknown>
  approverId?: string
  approverName?: string
  rejectReason?: string
  canResolve: boolean
  createdAt: string
  resolvedAt?: string
}

export interface KeyApprovalMeta {
  reason: string
  requestedBudget: number
  requestedModels: string[]
  departmentId: string
  departmentName: string
}

export interface MemberBudgetApprovalMeta {
  amount: number
  reason: string
  departmentId: string
  departmentName: string
}

export interface ProjectBudgetApprovalMeta {
  projectId: string
  projectName: string
  amount: number
  reason: string
}

export interface ProjectMemberBudgetApprovalMeta {
  projectId: string
  projectName: string
  amount: number
  reason: string
}

export interface ApprovalPreCheck {
  sufficient: boolean
  reservedPool: number
  requested: number
}

export interface ApprovalListResponse {
  items: ApprovalRequest[]
  total: number
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected' | 'cancelled' | 'failed'
export type ApprovalType = 'key' | 'member_budget'

export interface ApprovalRequest {
  id: string
  type: ApprovalType
  status: ApprovalStatus
  applicantId: string
  applicantName: string
  departmentId?: string
  departmentName?: string
  metadata: Record<string, unknown>
  approverId?: string
  approverName?: string
  rejectReason?: string
  createdAt: string
  resolvedAt?: string
}

export interface KeyApprovalMeta {
  reason: string
  requestedBudget: number
  requestedModels: string[]
}

export interface MemberBudgetApprovalMeta {
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

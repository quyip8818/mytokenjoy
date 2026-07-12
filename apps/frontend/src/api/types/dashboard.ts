import type { ProviderType } from './keys'

export interface CostSummary {
  totalCost: number
  totalCostMom: number
  totalTokens: number
  totalRequests: number
  totalRequestsMom: number
  avgCostPerRequest: number
  avgCostPerRequestMom: number
  avgCostPerMember: number
  avgCostPerMemberMom: number
}

export type CostPeriod =
  | 'current_month'
  | 'current_week'
  | 'last_month'
  | 'last_7_days'
  | 'last_30_days'
  | 'custom'

export type CostGranularity = 'day' | 'hour' | 'week' | 'month'

export interface CostQueryParams {
  period?: CostPeriod
  startDate?: string
  endDate?: string
  granularity?: CostGranularity
}

export interface DepartmentCost {
  departmentId: string
  departmentName: string
  cost: number
  percentage: number
  hasChildren?: boolean
}

export interface DepartmentCostMember {
  memberId: string
  memberName: string
  cost: number
  requests: number
  tokens: number
}

export interface DailyCost {
  date: string
  cost: number
  tokens: number
  requests: number
}

export interface TopConsumer {
  memberId: string
  memberName: string
  department: string
  cost: number
  tokens: number
  requests: number
}

export interface ModelUsage {
  callType: string
  modelId?: number
  modelName: string
  provider: ProviderType
  requests: number
  tokens: number
  cost: number
  percentage: number
}

export interface TeamUsage {
  departmentId: string
  departmentName: string
  budget: number
  consumed: number
  memberCount: number
  topModel: string
}

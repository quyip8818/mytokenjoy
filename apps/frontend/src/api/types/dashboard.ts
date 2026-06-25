import type { ProviderType } from './keys'

export interface CostSummary {
  totalCost: number
  monthOverMonth: number
  totalTokens: number
  totalRequests: number
  avgCostPerRequest: number
  avgCostPerMember: number
}

export type CostPeriod = 'current_month' | 'last_month' | 'last_7_days'

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
  modelId: string
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
  quota: number
  consumed: number
  memberCount: number
  topModel: string
}

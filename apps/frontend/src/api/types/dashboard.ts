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

export type CostPeriod = 'current_month' | 'last_month' | 'last_7_days' | 'custom'

export type CostGranularity = 'day' | 'hour' | 'week' | 'month'

export type UsageGranularity = 'day' | 'hour' | 'minute'

export type UsageSeriesGroupBy = 'none' | 'department' | 'member' | 'model'

export type UsageSeriesSource = 'buckets' | 'logs'

export type UsageMappingAsOf = 'ingest_time' | 'query_time'

export interface CostQueryParams {
  period?: CostPeriod
  startDate?: string
  endDate?: string
  granularity?: CostGranularity
}

export interface UsageSeriesQuery {
  granularity: UsageGranularity
  start: string
  end: string
  groupBy?: UsageSeriesGroupBy
  departmentId?: string
  memberId?: string
}

export interface UsageSeriesPoint {
  bucket: string
  departmentId?: string
  memberId?: string
  model?: string
  costCny: number
  callCount: number
  inputTokens: number
  outputTokens: number
}

export interface UsageSeriesResponse {
  granularity: UsageGranularity
  source: UsageSeriesSource
  timezone: string
  approximate: boolean
  mappingAsOf: UsageMappingAsOf
  unmappedCount?: number
  truncated?: boolean
  points: UsageSeriesPoint[]
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

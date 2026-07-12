export interface AccountStats {
  budgetRemaining: number
  totalSpent: number
}

export interface UsageStats {
  requestCount: number
  totalCount: number
}

export interface ResourceConsumption {
  totalCost: number
  totalTokens: number
}

export interface PerformanceStats {
  avgRPM: number
  avgTPM: number
}

export interface TimeSeriesPoint {
  time: string
  value: number
}

export interface NamedValue {
  name: string
  value: number
}

export interface ModelRank {
  model: string
  count: number
}

export interface MemberDashboardView {
  account: AccountStats
  usageStats: UsageStats
  resourceConsumption: ResourceConsumption
  performance: PerformanceStats
  consumptionTrend: TimeSeriesPoint[]
  consumptionDistribution: TimeSeriesPoint[]
  callDistribution: NamedValue[]
  callRanking: ModelRank[]
}

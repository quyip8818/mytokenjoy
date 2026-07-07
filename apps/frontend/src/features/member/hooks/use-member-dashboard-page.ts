import type { AppApis } from '@/api/app-apis'
import type {
  MemberDashboardView,
  ModelRank,
  NamedValue,
  TimeSeriesPoint,
} from '@/api/types/member'
import { queryKeys, useInjectedQuery } from '@/features/query'

const EMPTY_DASHBOARD: MemberDashboardView = {
  account: { balance: 0, totalSpent: 0 },
  usageStats: { requestCount: 0, totalCount: 0 },
  resourceConsumption: { totalCost: 0, totalTokens: 0 },
  performance: { avgRPM: 0, avgTPM: 0 },
  consumptionTrend: [],
  consumptionDistribution: [],
  callDistribution: [],
  callRanking: [],
}

export function useMemberDashboardPage(injectedApis?: AppApis) {
  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.member.dashboard(),
    queryFn: (api) => api.meApi.getDashboard(),
  })

  const dashboard = data ?? EMPTY_DASHBOARD

  return {
    loading,
    error,
    refresh,
    accountData: dashboard.account,
    usageStats: dashboard.usageStats,
    resourceConsumption: dashboard.resourceConsumption,
    performance: dashboard.performance,
    consumptionTrend: dashboard.consumptionTrend,
    consumptionDistribution: dashboard.consumptionDistribution,
    callDistribution: dashboard.callDistribution as NamedValue[],
    callRanking: dashboard.callRanking as ModelRank[],
    distributionTotal: sumSeries(dashboard.consumptionDistribution),
    trendTotal: sumSeries(dashboard.consumptionTrend),
    callTotal: dashboard.callDistribution.reduce((sum, item) => sum + item.value, 0),
  }
}

function sumSeries(points: TimeSeriesPoint[]) {
  return points.reduce((sum, point) => sum + point.value, 0)
}

import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type {
  CostGranularity,
  CostPeriod,
  CostQueryParams,
  DepartmentCost,
  DepartmentCostMember,
} from '@/api/types'
import { COST_GRANULARITY, COST_PERIOD } from '@/features/dashboard/lib/constants'
import { getMonthStartLocal, getTodayLocal } from '@/lib/date'
import { queryKeys, useInjectedQuery } from '@/features/query'
import {
  ROOT_DRILL,
  type DrillState,
  buildCostStats,
  buildDeptCostsWithColors,
  canDrillBack,
  drillBack,
  drillIntoDepartment,
  getDrillTitle,
} from '@/features/dashboard/lib/dashboard'

export type { CostStatItem, DrillState } from '@/features/dashboard/lib/dashboard'
export { ROOT_DRILL, COST_CHART_COLORS } from '@/features/dashboard/lib/dashboard'

function buildCostQuery(period: CostPeriod, startDate: string, endDate: string): CostQueryParams {
  if (period === COST_PERIOD.CUSTOM) {
    return { period, startDate, endDate }
  }
  return { period }
}

export function useCostDashboardPage(injectedApis?: AppApis) {
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [startDate, setStartDate] = useState(getMonthStartLocal)
  const [endDate, setEndDate] = useState(getTodayLocal)
  const [granularity, setGranularity] = useState<CostGranularity>(COST_GRANULARITY.DAY)
  const [drill, setDrill] = useState<DrillState>(ROOT_DRILL)

  const costQuery = useMemo(
    () => buildCostQuery(period, startDate, endDate),
    [period, startDate, endDate],
  )

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.dashboard.cost(costQuery, drill, granularity),
    queryFn: async (apis) => {
      const [summary, dailyCosts, deptCosts, topConsumers] = await Promise.all([
        apis.dashboardApi.getCostSummary(costQuery),
        apis.dashboardApi.getDailyCosts({ ...costQuery, granularity }),
        drill.level === 'members' && drill.deptId
          ? apis.dashboardApi.getDepartmentMemberCosts(drill.deptId, costQuery)
          : apis.dashboardApi.getDepartmentCosts({
              ...costQuery,
              parentId: drill.parentId ?? undefined,
            }),
        apis.dashboardApi.getTopConsumers({ ...costQuery, limit: 5 }),
      ])
      return { summary, dailyCosts, deptCosts, topConsumers }
    },
  })

  const handlePeriodChange = useCallback((value: string | null) => {
    if (!value) return
    setPeriod(value as CostPeriod)
    setDrill(ROOT_DRILL)
  }, [])

  const handleDrillDept = useCallback((dept: DepartmentCost) => {
    setDrill((current) => drillIntoDepartment(current, dept))
  }, [])

  const handleDrillBack = useCallback(() => {
    setDrill((current) => drillBack(current))
  }, [])

  const summary = data?.summary ?? null
  const dailyCosts = data?.dailyCosts ?? []
  const topConsumers = data?.topConsumers ?? []
  const rawDeptCosts = data?.deptCosts

  const deptCosts = useMemo(() => (rawDeptCosts ?? []) as DepartmentCost[], [rawDeptCosts])
  const memberCosts = useMemo(() => (rawDeptCosts ?? []) as DepartmentCostMember[], [rawDeptCosts])

  const deptCostsWithColors = useMemo(
    () => buildDeptCostsWithColors(drill.level, deptCosts, memberCosts),
    [deptCosts, memberCosts, drill.level],
  )

  const drillTitle = useMemo(() => getDrillTitle(drill), [drill])
  const stats = useMemo(() => buildCostStats(summary), [summary])
  const customDateInvalid =
    period === COST_PERIOD.CUSTOM && Boolean(startDate && endDate && startDate > endDate)

  return {
    period,
    startDate,
    endDate,
    granularity,
    customDateInvalid,
    drill,
    loading,
    error,
    refresh,
    summary,
    dailyCosts,
    topConsumers,
    deptCosts,
    memberCosts,
    deptCostsWithColors,
    drillTitle,
    stats,
    canDrillBack: canDrillBack(drill),
    handlePeriodChange,
    setStartDate,
    setEndDate,
    setGranularity,
    handleDrillDept,
    handleDrillBack,
  }
}

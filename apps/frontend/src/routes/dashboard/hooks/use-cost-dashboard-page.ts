import { useState, useMemo, useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'
import type { CostPeriod, DepartmentCost, DepartmentCostMember } from '@/api/types'
import { COST_PERIOD } from '@/lib/dashboard-constants'
import {
  ROOT_DRILL,
  type DrillState,
  buildCostStats,
  buildDeptCostsWithColors,
  canDrillBack,
  drillBack,
  drillIntoDepartment,
  getDrillTitle,
} from '@/lib/dashboard'

export type { CostStatItem, DrillState } from '@/lib/dashboard'
export { ROOT_DRILL, COST_CHART_COLORS } from '@/lib/dashboard'

export function useCostDashboardPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [drill, setDrill] = useState<DrillState>(ROOT_DRILL)

  const { data, loading, error, refresh } = useAsyncResource(async () => {
    const [summary, dailyCosts, deptCosts, topConsumers] = await Promise.all([
      apis.dashboardApi.getCostSummary(period),
      apis.dashboardApi.getDailyCosts(period),
      drill.level === 'members' && drill.deptId
        ? apis.dashboardApi.getDepartmentMemberCosts(drill.deptId, period)
        : apis.dashboardApi.getDepartmentCosts({
            parentId: drill.parentId ?? undefined,
            period,
          }),
      apis.dashboardApi.getTopConsumers({ limit: 5, period }),
    ])
    return { summary, dailyCosts, deptCosts, topConsumers }
  }, [apis, period, drill])

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

  return {
    period,
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
    handleDrillDept,
    handleDrillBack,
  }
}

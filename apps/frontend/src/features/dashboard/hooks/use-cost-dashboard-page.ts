import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { BudgetNode, CostGranularity, CostPeriod, CostQueryParams } from '@/api/types'
import { COST_GRANULARITY, COST_PERIOD } from '../lib/constants'
import { formatLocalDate, getCurrentBudgetPeriod, getMonthStartLocal, getTodayLocal, getWeekStartLocal } from '@/lib/date'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { buildCostStats, buildDeptCostsWithColors, COST_CHART_COLORS } from '../lib/dashboard'
import type { CostStatItem } from '../lib/dashboard'
import { budgetKeys, findBudgetNode } from '@/features/budget'

export type { CostStatItem }
export { COST_CHART_COLORS }

interface UseCostDashboardPageOptions {
  deptId: string | null
  injectedApis?: AppApis
}

function buildCostQuery(period: CostPeriod, startDate: string, endDate: string): CostQueryParams {
  if (period === COST_PERIOD.LAST_30_DAYS) {
    const to = new Date()
    const from = new Date()
    from.setDate(from.getDate() - 29)
    return { period: 'custom', startDate: formatLocalDate(from), endDate: formatLocalDate(to) }
  }
  if (period === COST_PERIOD.CURRENT_WEEK) {
    return { period: 'custom', startDate: getWeekStartLocal(), endDate: getTodayLocal() }
  }
  if (period === COST_PERIOD.CUSTOM) {
    return { period, startDate, endDate }
  }
  return { period }
}

export function useCostDashboardPage({ deptId, injectedApis }: UseCostDashboardPageOptions) {
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [startDate, setStartDate] = useState(getMonthStartLocal)
  const [endDate, setEndDate] = useState(getTodayLocal)
  const [granularity, setGranularity] = useState<CostGranularity>(COST_GRANULARITY.DAY)

  const costQuery = useMemo(
    () => buildCostQuery(period, startDate, endDate),
    [period, startDate, endDate],
  )

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.dashboard.cost(costQuery, deptId, granularity),
    queryFn: async (apis) => {
      const deptFilter = deptId ?? undefined
      const [summary, dailyCosts, deptCosts, topConsumers] = await Promise.all([
        apis.dashboardApi.getCostSummary({ ...costQuery, departmentId: deptFilter }),
        apis.dashboardApi.getDailyCosts({ ...costQuery, granularity, departmentId: deptFilter }),
        apis.dashboardApi.getDepartmentCosts({
          ...costQuery,
          parentId: deptId ?? undefined,
        }),
        apis.dashboardApi.getTopConsumers({ ...costQuery, limit: 5, departmentId: deptFilter }),
      ])
      return { summary, dailyCosts, deptCosts, topConsumers }
    },
  })

  const budgetPeriod = getCurrentBudgetPeriod()
  const { data: budgetTree = [], loading: budgetLoading } = useInjectedQuery({
    injectedApis,
    queryKey: budgetKeys.tree(budgetPeriod),
    queryFn: (apis) => apis.budgetApi.getTree(budgetPeriod),
  })

  const budgetSummary = useMemo(() => {
    if (budgetTree.length === 0) return { budget: 0, consumed: 0 }
    if (!deptId) {
      // Company-wide: sum all root nodes
      const budget = budgetTree.reduce((sum: number, n: BudgetNode) => sum + n.budget, 0)
      const consumed = budgetTree.reduce((sum: number, n: BudgetNode) => sum + n.consumed, 0)
      return { budget, consumed }
    }
    const node = findBudgetNode(budgetTree, deptId)
    return { budget: node?.budget ?? 0, consumed: node?.consumed ?? 0 }
  }, [budgetTree, deptId])

  const handlePeriodChange = useCallback((value: string | null) => {
    if (!value) return
    setPeriod(value as CostPeriod)
  }, [])

  const summary = data?.summary ?? null
  const dailyCosts = data?.dailyCosts ?? []
  const topConsumers = data?.topConsumers ?? []
  const deptCosts = useMemo(() => data?.deptCosts ?? [], [data?.deptCosts])

  const deptCostsWithColors = useMemo(
    () => buildDeptCostsWithColors('departments', deptCosts, []),
    [deptCosts],
  )

  const stats = useMemo(() => buildCostStats(summary), [summary])
  const customDateInvalid =
    period === COST_PERIOD.CUSTOM && Boolean(startDate && endDate && startDate > endDate)

  return {
    period,
    startDate,
    endDate,
    granularity,
    customDateInvalid,
    deptId,
    loading,
    budgetLoading,
    budgetSummary,
    error,
    refresh,
    summary,
    dailyCosts,
    topConsumers,
    deptCosts,
    deptCostsWithColors,
    stats,
    handlePeriodChange,
    setStartDate,
    setEndDate,
    setGranularity,
  }
}

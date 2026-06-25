import { useState, useMemo, useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'
import type { CostPeriod, DepartmentCost, DepartmentCostMember } from '@/api/types'
import { COST_PERIOD } from '@/lib/dashboard-constants'
import {
  TrendingUp,
  TrendingDown,
  Coins,
  Hash,
  Zap,
  DollarSign,
  User,
  type LucideIcon,
} from 'lucide-react'

export const COST_CHART_COLORS = ['#2563eb', '#3b82f6', '#10b981', '#f59e0b', '#06b6d4']

type DrillLevel = 'departments' | 'members'

export interface DrillState {
  level: DrillLevel
  parentId: string | null
  parentName: string | null
  deptId: string | null
  deptName: string | null
}

export const ROOT_DRILL: DrillState = {
  level: 'departments',
  parentId: null,
  parentName: null,
  deptId: null,
  deptName: null,
}

export interface CostStatItem {
  label: string
  value: string
  icon: LucideIcon
  accent: string
}

export function useCostDashboard(injectedApis?: AppApis) {
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

  const handleDrillDept = useCallback(
    (dept: DepartmentCost) => {
      if (drill.level === 'departments' && dept.hasChildren) {
        setDrill({
          level: 'departments',
          parentId: dept.departmentId,
          parentName: dept.departmentName,
          deptId: null,
          deptName: null,
        })
        return
      }
      if (drill.level === 'departments') {
        setDrill({
          level: 'members',
          parentId: drill.parentId,
          parentName: drill.parentName,
          deptId: dept.departmentId,
          deptName: dept.departmentName,
        })
      }
    },
    [drill],
  )

  const handleDrillBack = useCallback(() => {
    if (drill.level === 'members') {
      setDrill({
        level: 'departments',
        parentId: drill.parentId,
        parentName: drill.parentName,
        deptId: null,
        deptName: null,
      })
      return
    }
    if (drill.parentId) {
      setDrill(ROOT_DRILL)
    }
  }, [drill])

  const summary = data?.summary ?? null
  const dailyCosts = data?.dailyCosts ?? []
  const topConsumers = data?.topConsumers ?? []
  const rawDeptCosts = data?.deptCosts

  const deptCosts = useMemo(() => (rawDeptCosts ?? []) as DepartmentCost[], [rawDeptCosts])
  const memberCosts = useMemo(() => (rawDeptCosts ?? []) as DepartmentCostMember[], [rawDeptCosts])

  const deptCostsWithColors = useMemo(() => {
    if (drill.level === 'members') {
      return memberCosts.map((item, i) => ({
        departmentName: item.memberName,
        cost: item.cost,
        fill: COST_CHART_COLORS[i % COST_CHART_COLORS.length],
      }))
    }
    return deptCosts.map((item, i) => ({
      ...item,
      fill: COST_CHART_COLORS[i % COST_CHART_COLORS.length],
    }))
  }, [deptCosts, memberCosts, drill.level])

  const drillTitle = useMemo(() => {
    if (drill.level === 'members' && drill.deptName) return `${drill.deptName} · 成员明细`
    if (drill.parentName) return `${drill.parentName} · 子部门`
    return '部门花费明细'
  }, [drill])

  const stats: CostStatItem[] = [
    {
      label: '总花费',
      value: summary ? `¥${summary.totalCost.toLocaleString()}` : '-',
      icon: Coins,
      accent: 'from-blue-500 to-sky-500',
    },
    {
      label: '环比变化',
      value: summary ? `${summary.monthOverMonth > 0 ? '+' : ''}${summary.monthOverMonth}%` : '-',
      icon: summary && summary.monthOverMonth > 0 ? TrendingUp : TrendingDown,
      accent:
        summary && summary.monthOverMonth > 0
          ? 'from-red-400 to-rose-500'
          : 'from-emerald-400 to-teal-500',
    },
    {
      label: '人均成本',
      value: summary ? `¥${summary.avgCostPerMember.toLocaleString()}` : '-',
      icon: User,
      accent: 'from-violet-400 to-purple-500',
    },
    {
      label: '总调用次数',
      value: summary?.totalRequests.toLocaleString() ?? '-',
      icon: Zap,
      accent: 'from-amber-400 to-orange-500',
    },
    {
      label: '平均单次成本',
      value: summary ? `¥${summary.avgCostPerRequest.toFixed(2)}` : '-',
      icon: DollarSign,
      accent: 'from-cyan-400 to-blue-500',
    },
    {
      label: '总 Token',
      value: summary ? `${(summary.totalTokens / 1000000).toFixed(1)}M` : '-',
      icon: Hash,
      accent: 'from-blue-500 to-sky-400',
    },
  ]

  const canDrillBack = Boolean(drill.parentId || drill.level === 'members')

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
    canDrillBack,
    handlePeriodChange,
    handleDrillDept,
    handleDrillBack,
  }
}

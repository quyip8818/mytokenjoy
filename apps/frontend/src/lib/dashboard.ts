import type { CostSummary, DepartmentCost, DepartmentCostMember } from '@/api/types'
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

export type DrillLevel = 'departments' | 'members'

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

export function drillIntoDepartment(drill: DrillState, dept: DepartmentCost): DrillState {
  if (drill.level === 'departments' && dept.hasChildren) {
    return {
      level: 'departments',
      parentId: dept.departmentId,
      parentName: dept.departmentName,
      deptId: null,
      deptName: null,
    }
  }
  if (drill.level === 'departments') {
    return {
      level: 'members',
      parentId: drill.parentId,
      parentName: drill.parentName,
      deptId: dept.departmentId,
      deptName: dept.departmentName,
    }
  }
  return drill
}

export function drillBack(drill: DrillState): DrillState {
  if (drill.level === 'members') {
    return {
      level: 'departments',
      parentId: drill.parentId,
      parentName: drill.parentName,
      deptId: null,
      deptName: null,
    }
  }
  if (drill.parentId) {
    return ROOT_DRILL
  }
  return drill
}

export function getDrillTitle(drill: DrillState): string {
  if (drill.level === 'members' && drill.deptName) return `${drill.deptName} · 成员明细`
  if (drill.parentName) return `${drill.parentName} · 子部门`
  return '部门花费明细'
}

export function canDrillBack(drill: DrillState): boolean {
  return Boolean(drill.parentId || drill.level === 'members')
}

export function buildDeptCostsWithColors(
  drillLevel: DrillLevel,
  deptCosts: DepartmentCost[],
  memberCosts: DepartmentCostMember[],
) {
  if (drillLevel === 'members') {
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
}

export function buildCostStats(summary: CostSummary | null): CostStatItem[] {
  return [
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
}

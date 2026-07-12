import type { CostSummary, DepartmentCost, DepartmentCostMember } from '@/api/types'
import { Coins, Zap, User, type LucideIcon } from 'lucide-react'

export const COST_CHART_COLORS = ['#4f46e5', '#7c3aed', '#10b981', '#f59e0b', '#06b6d4']

export interface CostStatItem {
  label: string
  value: string
  mom?: number
  icon: LucideIcon
  accent: string
}

function formatMom(mom: number): string {
  const rounded = Number(mom.toFixed(2))
  return `${rounded > 0 ? '+' : ''}${rounded}%`
}

export function formatTokenCount(tokens: number): string {
  if (tokens <= 0) return '0'
  return `${(tokens / 1000000).toFixed(1)}M`
}

export function buildDeptCostsWithColors(
  drillLevel: 'departments' | 'members',
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
      value: summary ? `¥${summary.totalCost.toFixed(2)}` : '-',
      mom: summary?.totalCostMom,
      icon: Coins,
      accent: 'bg-primary',
    },
    {
      label: '总调用次数',
      value: summary?.totalRequests.toLocaleString() ?? '-',
      mom: summary?.totalRequestsMom,
      icon: Zap,
      accent: 'bg-amber-400',
    },
    {
      label: '人均成本',
      value: summary ? `¥${summary.avgCostPerMember.toFixed(2)}` : '-',
      mom: summary?.avgCostPerMemberMom,
      icon: User,
      accent: 'bg-violet-500',
    },
  ]
}

export { formatMom }

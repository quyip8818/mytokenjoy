import type {
  CostSummary,
  DepartmentCost,
  DepartmentCostMember,
  UsageGranularity,
  UsageSeriesPoint,
} from '@/api/types'
import { demoSeriesAnchorEnd, demoSeriesMonthStartISO } from './demo-series'
import { Coins, Hash, Zap, DollarSign, User, type LucideIcon } from 'lucide-react'

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
  if (tokens <= 0) return '-'
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
      label: '平均单次成本',
      value: summary ? `¥${summary.avgCostPerRequest.toFixed(2)}` : '-',
      mom: summary?.avgCostPerRequestMom,
      icon: DollarSign,
      accent: 'bg-cyan-400',
    },
    {
      label: '人均成本',
      value: summary ? `¥${summary.avgCostPerMember.toFixed(2)}` : '-',
      mom: summary?.avgCostPerMemberMom,
      icon: User,
      accent: 'bg-violet-500',
    },
    {
      label: '总调用次数',
      value: summary?.totalRequests.toLocaleString() ?? '-',
      mom: summary?.totalRequestsMom,
      icon: Zap,
      accent: 'bg-amber-400',
    },
    {
      label: '总 Token',
      value: summary ? formatTokenCount(summary.totalTokens) : '-',
      icon: Hash,
      accent: 'bg-violet-500',
    },
  ]
}

export { formatMom }

export interface UsageSeriesChartPoint {
  bucket: string
  label: string
  costCny: number
  callCount: number
}

function formatUsageSeriesBucketLabel(bucket: string, granularity: UsageGranularity): string {
  const parsed = new Date(bucket)
  if (!Number.isNaN(parsed.getTime())) {
    const hours = String(parsed.getHours()).padStart(2, '0')
    const minutes = String(parsed.getMinutes()).padStart(2, '0')
    if (granularity === 'minute') {
      return `${hours}:${minutes}`
    }
    const month = String(parsed.getMonth() + 1).padStart(2, '0')
    const day = String(parsed.getDate()).padStart(2, '0')
    return `${month}-${day} ${hours}:00`
  }
  if (granularity === 'minute' && bucket.length >= 16) {
    return bucket.slice(11, 16)
  }
  if (bucket.length >= 16) {
    return bucket.slice(5, 16)
  }
  return bucket
}

export function buildUsageSeriesChartData(
  points: UsageSeriesPoint[],
  granularity: UsageGranularity,
): UsageSeriesChartPoint[] {
  const byBucket = new Map<string, { costCny: number; callCount: number }>()
  for (const point of points) {
    const existing = byBucket.get(point.bucket) ?? { costCny: 0, callCount: 0 }
    byBucket.set(point.bucket, {
      costCny: existing.costCny + point.costCny,
      callCount: existing.callCount + point.callCount,
    })
  }
  return [...byBucket.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([bucket, values]) => ({
      bucket,
      label: formatUsageSeriesBucketLabel(bucket, granularity),
      costCny: values.costCny,
      callCount: values.callCount,
    }))
}

export function buildUsageSeriesWindow(granularity: UsageGranularity): {
  start: string
  end: string
} {
  const end = import.meta.env.DEV ? demoSeriesAnchorEnd() : new Date()
  if (granularity === 'minute') {
    const start = new Date(end)
    start.setTime(end.getTime() - 3 * 60 * 60 * 1000)
    return { start: start.toISOString(), end: end.toISOString() }
  }
  if (import.meta.env.DEV) {
    return { start: demoSeriesMonthStartISO(), end: end.toISOString() }
  }
  const start = new Date(end)
  start.setHours(start.getHours() - 24)
  return { start: start.toISOString(), end: end.toISOString() }
}

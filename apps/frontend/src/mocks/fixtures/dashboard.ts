import type {
  CostPeriod,
  CostSummary,
  DailyCost,
  DepartmentCost,
  DepartmentCostMember,
  ModelUsage,
  TeamUsage,
  TopConsumer,
} from '@/api/types'
import { mockMembers } from './org'

const DAILY_COST_FACTORS = [
  0.82, 0.91, 0.88, 0.95, 1.02, 0.97, 0.85, 0.93, 1.05, 1.1, 0.98, 1.03, 0.89, 0.94, 1.08, 1.12,
  0.96, 1.0, 0.87, 0.92, 1.06, 1.15, 0.99, 1.04, 0.9, 0.86, 1.01, 1.09, 0.95, 1.07,
]

const PERIOD_SCALE: Record<CostPeriod, number> = {
  current_month: 1,
  last_month: 0.88,
  last_7_days: 0.28,
}

const PERIOD_MOM: Record<CostPeriod, number> = {
  current_month: 12.5,
  last_month: 8.2,
  last_7_days: -3.1,
}

const TOTAL_COST_TARGET = 67500

export function buildCostSummary(period: CostPeriod = 'current_month'): CostSummary {
  const scale = PERIOD_SCALE[period]
  const totalCost = Math.round(TOTAL_COST_TARGET * scale)
  const memberCount = mockMembers.filter((m) => m.status === 'active').length
  const totalRequests = Math.round(28500 * scale)
  return {
    totalCost,
    monthOverMonth: PERIOD_MOM[period],
    totalTokens: Math.round(45000000 * scale),
    totalRequests,
    avgCostPerRequest: totalRequests > 0 ? Math.round((totalCost / totalRequests) * 100) / 100 : 0,
    avgCostPerMember: memberCount > 0 ? Math.round(totalCost / memberCount) : 0,
  }
}

export const mockCostSummary: CostSummary = buildCostSummary('current_month')

const TOP_LEVEL_DEPARTMENTS: Omit<DepartmentCost, 'percentage'>[] = [
  { departmentId: 'dept-2', departmentName: '技术部', cost: 38200, hasChildren: true },
  { departmentId: 'dept-6', departmentName: '产品部', cost: 14300, hasChildren: false },
  { departmentId: 'dept-7', departmentName: '市场部', cost: 8500, hasChildren: false },
  { departmentId: 'dept-8', departmentName: '行政部', cost: 6500, hasChildren: false },
]

const CHILD_DEPARTMENTS: Record<string, Omit<DepartmentCost, 'percentage'>[]> = {
  'dept-2': [
    { departmentId: 'dept-3', departmentName: '后端组', cost: 21000, hasChildren: true },
    { departmentId: 'dept-4', departmentName: '前端组', cost: 11200, hasChildren: true },
    { departmentId: 'dept-5', departmentName: '测试组', cost: 6000, hasChildren: true },
  ],
}

const DEPT_MEMBER_COSTS: Record<string, DepartmentCostMember[]> = {
  'dept-3': [
    { memberId: 'm-2', memberName: '李四', cost: 12500, requests: 5200, tokens: 8500000 },
    { memberId: 'm-1', memberName: '张三', cost: 8700, requests: 3800, tokens: 5800000 },
  ],
  'dept-4': [
    { memberId: 'm-4', memberName: '赵六', cost: 6200, requests: 2900, tokens: 4100000 },
    { memberId: 'm-5', memberName: '钱七', cost: 5000, requests: 2200, tokens: 3200000 },
  ],
  'dept-5': [
    { memberId: 'm-6', memberName: '孙八', cost: 3500, requests: 1500, tokens: 2300000 },
    { memberId: 'm-7', memberName: '周九', cost: 2500, requests: 1100, tokens: 1800000 },
  ],
}

export function getDepartmentCostsForParent(
  parentId: string | null,
  period: CostPeriod = 'current_month',
): DepartmentCost[] {
  const scale = PERIOD_SCALE[period]
  const rows = parentId === null ? TOP_LEVEL_DEPARTMENTS : (CHILD_DEPARTMENTS[parentId] ?? [])
  const total = rows.reduce((sum, r) => sum + r.cost, 0) || 1
  return rows.map((row) => ({
    ...row,
    cost: Math.round(row.cost * scale),
    percentage: Math.round((row.cost / total) * 1000) / 10,
  }))
}

export function getDepartmentMemberCosts(
  deptId: string,
  period: CostPeriod = 'current_month',
): DepartmentCostMember[] {
  const scale = PERIOD_SCALE[period]
  return (DEPT_MEMBER_COSTS[deptId] ?? []).map((row) => ({
    ...row,
    cost: Math.round(row.cost * scale),
    requests: Math.round(row.requests * scale),
    tokens: Math.round(row.tokens * scale),
  }))
}

function buildDailyCosts(period: CostPeriod = 'current_month'): DailyCost[] {
  const scale = PERIOD_SCALE[period]
  const factors = period === 'last_7_days' ? DAILY_COST_FACTORS.slice(-7) : DAILY_COST_FACTORS
  const factorSum = factors.reduce((sum, f) => sum + f, 0)
  const baseDaily = (TOTAL_COST_TARGET * scale) / factorSum
  const startDay = period === 'last_7_days' ? 24 : 1

  return factors.map((factor, i) => {
    const date = new Date(2026, 5, startDay + i)
    const cost = Math.round(baseDaily * factor * 100) / 100
    return {
      date: date.toISOString().split('T')[0],
      cost,
      tokens: Math.round(cost * 700),
      requests: Math.round(cost / 2.37),
    }
  })
}

export const mockDailyCosts: DailyCost[] = buildDailyCosts('current_month')

export const mockDepartmentCosts: DepartmentCost[] = getDepartmentCostsForParent(null)

const TOP_CONSUMER_SPECS: { memberId: string; cost: number; tokens: number; requests: number }[] = [
  { memberId: 'm-2', cost: 12500, tokens: 8500000, requests: 5200 },
  { memberId: 'm-1', cost: 8700, tokens: 5800000, requests: 3800 },
  { memberId: 'm-4', cost: 6200, tokens: 4100000, requests: 2900 },
  { memberId: 'm-11', cost: 5400, tokens: 3600000, requests: 2500 },
  { memberId: 'm-14', cost: 4800, tokens: 3200000, requests: 2200 },
  { memberId: 'm-5', cost: 4200, tokens: 2800000, requests: 1900 },
  { memberId: 'm-18', cost: 3900, tokens: 2600000, requests: 1700 },
  { memberId: 'm-22', cost: 3500, tokens: 2300000, requests: 1500 },
  { memberId: 'm-3', cost: 3200, tokens: 2100000, requests: 1400 },
  { memberId: 'm-25', cost: 2800, tokens: 1800000, requests: 1200 },
]

export function getTopConsumers(limit = 5, period: CostPeriod = 'current_month') {
  const scale = PERIOD_SCALE[period]
  return TOP_CONSUMER_SPECS.slice(0, limit).map((spec) => {
    const member = mockMembers.find((m) => m.id === spec.memberId)
    return {
      memberId: spec.memberId,
      memberName: member?.name ?? spec.memberId,
      department: member?.departmentName ?? '',
      cost: Math.round(spec.cost * scale),
      tokens: Math.round(spec.tokens * scale),
      requests: Math.round(spec.requests * scale),
    }
  })
}

export const mockTopConsumers: TopConsumer[] = getTopConsumers()

export const mockModelUsage: ModelUsage[] = [
  {
    modelId: 'model-1',
    modelName: 'GPT-4o',
    provider: 'openai',
    requests: 12000,
    tokens: 18000000,
    cost: 32000,
    percentage: 47.4,
  },
  {
    modelId: 'model-5',
    modelName: 'DeepSeek V3',
    provider: 'deepseek',
    requests: 8500,
    tokens: 15000000,
    cost: 12500,
    percentage: 18.5,
  },
  {
    modelId: 'model-4',
    modelName: 'Claude Sonnet 4.6',
    provider: 'anthropic',
    requests: 4500,
    tokens: 7000000,
    cost: 14000,
    percentage: 20.7,
  },
  {
    modelId: 'model-2',
    modelName: 'GPT-4o Mini',
    provider: 'openai',
    requests: 3000,
    tokens: 4000000,
    cost: 5500,
    percentage: 8.1,
  },
  {
    modelId: 'model-8',
    modelName: 'Qwen Plus',
    provider: 'qwen',
    requests: 500,
    tokens: 1000000,
    cost: 3500,
    percentage: 5.2,
  },
]

export const mockTeamUsage: TeamUsage[] = [
  {
    departmentId: 'dept-3',
    departmentName: '后端组',
    quota: 25000,
    consumed: 21000,
    memberCount: 20,
    topModel: 'GPT-4o',
  },
  {
    departmentId: 'dept-4',
    departmentName: '前端组',
    quota: 15000,
    consumed: 11200,
    memberCount: 15,
    topModel: 'Claude Sonnet 4.6',
  },
  {
    departmentId: 'dept-5',
    departmentName: '测试组',
    quota: 10000,
    consumed: 6000,
    memberCount: 10,
    topModel: 'DeepSeek V3',
  },
  {
    departmentId: 'dept-6',
    departmentName: '产品部',
    quota: 20000,
    consumed: 14300,
    memberCount: 25,
    topModel: 'GPT-4o Mini',
  },
  {
    departmentId: 'dept-7',
    departmentName: '市场部',
    quota: 15000,
    consumed: 8500,
    memberCount: 30,
    topModel: 'Qwen Plus',
  },
  {
    departmentId: 'dept-8',
    departmentName: '行政部',
    quota: 15000,
    consumed: 6500,
    memberCount: 28,
    topModel: 'GPT-4o Mini',
  },
]

export { buildDailyCosts }

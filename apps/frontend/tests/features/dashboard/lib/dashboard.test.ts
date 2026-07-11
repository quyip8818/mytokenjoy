import { describe, expect, it } from 'vitest'
import { buildCostStats, formatTokenCount } from '@/features/dashboard/lib/dashboard'
import type { CostSummary } from '@/api/types'

describe('buildCostStats', () => {
  it('maps each primary metric with its own mom field', () => {
    const summary: CostSummary = {
      totalCost: 1000,
      totalCostMom: 12.5,
      totalTokens: 1000000,
      totalRequests: 200,
      totalRequestsMom: 8,
      avgCostPerRequest: 5,
      avgCostPerRequestMom: -2,
      avgCostPerMember: 50,
      avgCostPerMemberMom: 4,
    }
    const stats = buildCostStats(summary)
    expect(stats.find((s) => s.label === '总花费')?.mom).toBe(12.5)
    expect(stats.find((s) => s.label === '平均单次成本')?.mom).toBe(-2)
    expect(stats.find((s) => s.label === '人均成本')?.mom).toBe(4)
    expect(stats.find((s) => s.label === '总调用次数')?.mom).toBe(8)
    expect(stats.find((s) => s.label === '总 Token')?.mom).toBeUndefined()
  })

  it('shows dash for zero total tokens', () => {
    const summary: CostSummary = {
      totalCost: 1000,
      totalCostMom: 0,
      totalTokens: 0,
      totalRequests: 0,
      totalRequestsMom: 0,
      avgCostPerRequest: 0,
      avgCostPerRequestMom: 0,
      avgCostPerMember: 0,
      avgCostPerMemberMom: 0,
    }
    const stats = buildCostStats(summary)
    expect(stats.find((s) => s.label === '总 Token')?.value).toBe('-')
    expect(formatTokenCount(0)).toBe('-')
  })
})

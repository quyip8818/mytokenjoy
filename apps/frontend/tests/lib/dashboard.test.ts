import { describe, expect, it } from 'vitest'
import { aggregateDailyCosts, buildCostStats } from '@/lib/dashboard'
import type { CostSummary, DailyCost } from '@/api/types'

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
})

describe('aggregateDailyCosts', () => {
  const daily: DailyCost[] = [
    { date: '2026-06-02', cost: 10, tokens: 100, requests: 5 },
    { date: '2026-06-03', cost: 20, tokens: 200, requests: 10 },
    { date: '2026-06-09', cost: 30, tokens: 300, requests: 15 },
  ]

  it('returns daily rows unchanged for day granularity', () => {
    expect(aggregateDailyCosts(daily, 'day')).toEqual(daily)
  })

  it('aggregates rows by month', () => {
    const result = aggregateDailyCosts(daily, 'month')
    expect(result).toHaveLength(1)
    expect(result[0].cost).toBe(60)
  })
})

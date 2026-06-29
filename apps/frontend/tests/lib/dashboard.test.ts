import { describe, expect, it } from 'vitest'
import {
  buildCostStats,
  buildUsageSeriesChartData,
  buildUsageSeriesWindow,
  formatTokenCount,
} from '@/lib/dashboard'
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

describe('buildUsageSeriesWindow', () => {
  it('anchors dev ranges to demo today seed window', () => {
    const hourWindow = buildUsageSeriesWindow('hour')
    expect(hourWindow.end).toContain('2026-06-19')
    expect(hourWindow.start).toContain('2026-06-01')

    const minuteWindow = buildUsageSeriesWindow('minute')
    expect(minuteWindow.end).toContain('2026-06-19')
    expect(new Date(minuteWindow.end).getTime() - new Date(minuteWindow.start).getTime()).toBe(
      3 * 60 * 60 * 1000,
    )
  })
})

describe('buildUsageSeriesChartData', () => {
  it('aggregates points by bucket and formats hour labels', () => {
    const chartData = buildUsageSeriesChartData(
      [
        {
          bucket: '2026-06-10T09:00:00+08:00',
          costCny: 4,
          callCount: 2,
          inputTokens: 0,
          outputTokens: 0,
        },
        {
          bucket: '2026-06-10T09:00:00+08:00',
          costCny: 6,
          callCount: 3,
          inputTokens: 0,
          outputTokens: 0,
        },
      ],
      'hour',
    )

    expect(chartData).toHaveLength(1)
    expect(chartData[0]?.costCny).toBe(10)
    expect(chartData[0]?.callCount).toBe(5)
    expect(chartData[0]?.label).toMatch(/06-10/)
  })
})

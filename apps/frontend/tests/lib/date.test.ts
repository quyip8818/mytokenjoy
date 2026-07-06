import { describe, expect, it, vi, afterEach } from 'vitest'
import {
  formatLocalDate,
  getMonthStartLocal,
  getTodayLocal,
  resolveLast7DaysRange,
} from '@/lib/date'

describe('date', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('formats local ISO date', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 19, 12, 0, 0))
    expect(formatLocalDate(new Date())).toBe('2026-06-19')
    expect(getTodayLocal()).toBe('2026-06-19')
    expect(getMonthStartLocal()).toBe('2026-06-01')
  })

  it('resolves last 7 days ending today', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 19, 12, 0, 0))
    expect(resolveLast7DaysRange()).toEqual({ from: '2026-06-13', to: '2026-06-19' })
  })
})

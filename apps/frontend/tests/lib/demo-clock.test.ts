import { describe, expect, it } from 'vitest'
import { DEMO_MONTH_START, DEMO_TODAY, resolveLast7DaysRange } from '@/lib/demo-clock'

describe('demo-clock', () => {
  it('exposes fixed demo anchor dates', () => {
    expect(DEMO_TODAY).toBe('2026-06-19')
    expect(DEMO_MONTH_START).toBe('2026-06-01')
  })

  it('resolves last 7 days ending on demo today', () => {
    const range = resolveLast7DaysRange()
    expect(range.to).toBe(DEMO_TODAY)
    expect(range.from).toBe('2026-06-13')
  })
})

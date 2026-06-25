import { describe, expect, it } from 'vitest'
import { DEMO_MONTH_START, DEMO_TODAY } from '@/mocks/lib/demo-clock'

describe('mocks demo-clock', () => {
  it('exposes fixture anchor dates for MSW only', () => {
    expect(DEMO_TODAY).toBe('2026-06-19')
    expect(DEMO_MONTH_START).toBe('2026-06-01')
  })
})

import { describe, expect, it, vi, afterEach } from 'vitest'
import { AUDIT_DATE_PRESET } from '@/lib/audit-constants'
import { resolveLast7DaysRange } from '@/lib/date'
import {
  AUDIT_FILTER_ALL,
  buildAuditBaseQuery,
  buildCallsQuery,
  buildOperationsQuery,
  omitAll,
} from '@/lib/audit-query'

describe('audit-query', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('omitAll maps all sentinel to undefined', () => {
    expect(omitAll(AUDIT_FILTER_ALL)).toBeUndefined()
    expect(omitAll('success')).toBe('success')
  })

  it('buildAuditBaseQuery trims keyword and resolves date preset', () => {
    expect(
      buildAuditBaseQuery({
        datePreset: AUDIT_DATE_PRESET.ALL,
        keyword: '  bot  ',
      }),
    ).toEqual({ keyword: 'bot' })

    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 5, 19, 12, 0, 0))
    const range = resolveLast7DaysRange()
    expect(
      buildAuditBaseQuery({
        datePreset: AUDIT_DATE_PRESET.LAST_7_DAYS,
        keyword: '',
      }),
    ).toEqual({ ...range, keyword: undefined })
  })

  it('buildCallsQuery maps domain filters', () => {
    expect(
      buildCallsQuery({
        status: AUDIT_FILTER_ALL,
        callerId: 'm-1',
        datePreset: AUDIT_DATE_PRESET.ALL,
        keyword: '',
      }),
    ).toEqual({ callerId: 'm-1' })
  })

  it('buildOperationsQuery maps domain filters', () => {
    expect(
      buildOperationsQuery({
        action: 'key_create',
        operatorId: AUDIT_FILTER_ALL,
        datePreset: AUDIT_DATE_PRESET.ALL,
        keyword: '预算',
      }),
    ).toEqual({ action: 'key_create', keyword: '预算' })
  })
})

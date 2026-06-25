import { describe, expect, it } from 'vitest'
import { AUDIT_DATE_PRESET } from '@/lib/audit-constants'
import {
  AUDIT_FILTER_ALL,
  buildAuditBaseQuery,
  buildCallsQuery,
  buildOperationsQuery,
  omitAll,
} from '@/lib/audit-query'

describe('audit-query', () => {
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

    expect(
      buildAuditBaseQuery({
        datePreset: AUDIT_DATE_PRESET.LAST_7_DAYS,
        keyword: '',
      }),
    ).toEqual({ from: '2026-06-13', to: '2026-06-19', keyword: undefined })
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

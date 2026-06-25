import { resolveLast7DaysRange } from '@/lib/date'

export const AUDIT_DATE_PRESET = {
  ALL: 'all',
  LAST_7_DAYS: 'last_7_days',
} as const

export const AUDIT_DATE_PRESET_LABELS: Record<string, string> = {
  [AUDIT_DATE_PRESET.ALL]: '全部时间',
  [AUDIT_DATE_PRESET.LAST_7_DAYS]: '近 7 天',
}

export function resolveAuditDatePreset(preset: string): { from?: string; to?: string } {
  if (preset === AUDIT_DATE_PRESET.LAST_7_DAYS) {
    return resolveLast7DaysRange()
  }
  return {}
}

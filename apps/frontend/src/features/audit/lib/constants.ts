import { formatLocalDate } from '@/lib/date'

export const AUDIT_PAGE_SIZE = 20

export const AUDIT_DATE_PRESET = {
  ALL: 'all',
  TODAY: 'today',
  YESTERDAY: 'yesterday',
  LAST_7_DAYS: 'last_7_days',
  LAST_30_DAYS: 'last_30_days',
} as const

export const AUDIT_DATE_PRESET_LABELS: Record<string, string> = {
  [AUDIT_DATE_PRESET.ALL]: '全部时间',
  [AUDIT_DATE_PRESET.TODAY]: '今天',
  [AUDIT_DATE_PRESET.YESTERDAY]: '昨天',
  [AUDIT_DATE_PRESET.LAST_7_DAYS]: '最近一周',
  [AUDIT_DATE_PRESET.LAST_30_DAYS]: '最近一月',
}

export function resolveAuditDatePreset(preset: string): { from?: string; to?: string } {
  const now = new Date()
  const today = formatLocalDate(now)

  switch (preset) {
    case AUDIT_DATE_PRESET.TODAY:
      return { from: today, to: today }
    case AUDIT_DATE_PRESET.YESTERDAY: {
      const yesterday = new Date(now)
      yesterday.setDate(yesterday.getDate() - 1)
      return { from: formatLocalDate(yesterday), to: formatLocalDate(yesterday) }
    }
    case AUDIT_DATE_PRESET.LAST_7_DAYS: {
      const from = new Date(now)
      from.setDate(from.getDate() - 6)
      return { from: formatLocalDate(from), to: today }
    }
    case AUDIT_DATE_PRESET.LAST_30_DAYS: {
      const from = new Date(now)
      from.setDate(from.getDate() - 29)
      return { from: formatLocalDate(from), to: today }
    }
    default:
      return {}
  }
}

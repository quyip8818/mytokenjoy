import { getCurrentBudgetPeriod } from '@/lib/date'

export type PeriodState = 'past' | 'current' | 'future'

export function getPeriodState(period: string): PeriodState {
  const current = getCurrentBudgetPeriod()
  if (period < current) return 'past'
  if (period === current) return 'current'
  return 'future'
}

/** Only current and next month are editable; past months are frozen. */
export function isPeriodEditable(period: string): boolean {
  return getPeriodState(period) !== 'past'
}

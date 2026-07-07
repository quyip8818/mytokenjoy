import { getTodayLocal } from '@/lib/date'

export function demoSeriesAnchorEnd(): Date {
  const today = getTodayLocal()
  return new Date(`${today}T23:59:59+08:00`)
}

export function demoSeriesMonthStartISO(): string {
  const [year, month] = getTodayLocal().split('-')
  return `${year}-${month}-01T00:00:00+08:00`
}

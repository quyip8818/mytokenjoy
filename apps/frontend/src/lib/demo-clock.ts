export const DEMO_TODAY = '2026-06-19'

export function demoSeriesAnchorEnd(): Date {
  return new Date(`${DEMO_TODAY}T23:59:59+08:00`)
}

export function demoSeriesMonthStartISO(): string {
  const [year, month] = DEMO_TODAY.split('-')
  return `${year}-${month}-01T00:00:00+08:00`
}

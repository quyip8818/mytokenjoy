export const DEMO_TODAY = '2026-06-19'
export const DEMO_MONTH_START = '2026-06-01'

function parseDemoDate(isoDate: string): Date {
  const [year, month, day] = isoDate.split('-').map(Number)
  return new Date(year, month - 1, day)
}

function formatLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function resolveLast7DaysRange(): { from: string; to: string } {
  const from = parseDemoDate(DEMO_TODAY)
  from.setDate(from.getDate() - 6)
  return {
    from: formatLocalDate(from),
    to: DEMO_TODAY,
  }
}

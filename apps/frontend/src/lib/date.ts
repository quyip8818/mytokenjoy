export function formatLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function getTodayLocal(): string {
  return formatLocalDate(new Date())
}

export function getMonthStartLocal(): string {
  const now = new Date()
  return formatLocalDate(new Date(now.getFullYear(), now.getMonth(), 1))
}

export function getCurrentBudgetPeriod(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

export function resolveLast7DaysRange(): { from: string; to: string } {
  const to = new Date()
  const from = new Date()
  from.setDate(from.getDate() - 6)
  return {
    from: formatLocalDate(from),
    to: formatLocalDate(to),
  }
}

export function getWeekStartLocal(): string {
  const now = new Date()
  const day = now.getDay()
  // Monday = 1, Sunday = 0 → offset to get Monday as week start
  const diff = day === 0 ? 6 : day - 1
  const monday = new Date(now.getFullYear(), now.getMonth(), now.getDate() - diff)
  return formatLocalDate(monday)
}

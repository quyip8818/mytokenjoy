/**
 * ISO 时间戳 → "YYYY-MM-DD HH:mm:SS"（系统本地时区）
 * ponytail: 全站唯一时间格式化入口，不要在组件里自行 format
 */
export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  return `${formatLocalDate(d)} ${hours}:${minutes}:${seconds}`
}

/**
 * 相对时间：刚刚 / N分钟前 / N小时前 / N天前 / 绝对时间
 */
export function formatTimeAgo(iso: string): string {
  const date = new Date(iso)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / 60_000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `${diffHours} 小时前`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays} 天前`
  return formatDateTime(iso)
}

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

export function teamUsagePercent(consumed: number, quota: number): number {
  if (quota <= 0) return 0
  return Math.min(100, Math.round((consumed / quota) * 100))
}

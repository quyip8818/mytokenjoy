export function departmentUsagePercent(consumed: number, budget: number): number {
  if (budget <= 0) return 0
  return Math.min(100, Math.round((consumed / budget) * 100))
}

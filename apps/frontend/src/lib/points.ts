export const DEFAULT_POINTS_PER_UNIT = 1000

export function pointsToDisplay(points: number): number {
  return points / DEFAULT_POINTS_PER_UNIT
}

export function displayToPoints(display: number): number {
  const rounded = Math.round(display * 100) / 100
  return Math.round(rounded * DEFAULT_POINTS_PER_UNIT)
}

/** Format point amounts (budget / key limits). */
export function formatDisplayCurrency(points: number): string {
  return formatMoney(pointsToDisplay(points))
}

/** Format amounts already in display currency (CallLog.cost / wallet / dashboard Spend). */
export function formatMoney(amount: number): string {
  return `¥${Number(amount).toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`
}

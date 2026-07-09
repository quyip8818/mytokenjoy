export const DEFAULT_POINTS_PER_UNIT = 1000

export function pointsToDisplay(points: number): number {
  return points / DEFAULT_POINTS_PER_UNIT
}

export function displayToPoints(display: number): number {
  return display * DEFAULT_POINTS_PER_UNIT
}

export function formatDisplayCurrency(points: number): string {
  return `¥${pointsToDisplay(points).toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`
}

export const DEFAULT_POINTS_PER_UNIT = 1000

/** Build conversion helpers for a company PPU. */
export function createBillingExchange(pointsPerUnit: number = DEFAULT_POINTS_PER_UNIT) {
  const ppu = pointsPerUnit > 0 ? pointsPerUnit : DEFAULT_POINTS_PER_UNIT
  const pointsToDisplayFn = (points: number) => points / ppu
  const displayToPointsFn = (display: number) => {
    const rounded = Math.round(display * 100) / 100
    return Math.round(rounded * ppu)
  }
  return {
    pointsPerUnit: ppu,
    pointsToDisplay: pointsToDisplayFn,
    displayToPoints: displayToPointsFn,
    formatDisplayCurrency: (points: number) => formatMoney(pointsToDisplayFn(points)),
  }
}

export type BillingExchange = ReturnType<typeof createBillingExchange>

let active = createBillingExchange()

/** AuthSessionProvider syncs session PPU here; form imports use these module helpers. */
export function setActiveBillingExchange(exchange: BillingExchange): void {
  active = exchange
}

export function pointsToDisplay(points: number): number {
  return active.pointsToDisplay(points)
}

export function displayToPoints(display: number): number {
  return active.displayToPoints(display)
}

/** Format point amounts (budget / key limits). */
export function formatDisplayCurrency(points: number): string {
  return active.formatDisplayCurrency(points)
}

/** Format amounts already in display currency (CallLog.cost / wallet / dashboard Spend). */
export function formatMoney(amount: number): string {
  return `¥${Number(amount).toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`
}

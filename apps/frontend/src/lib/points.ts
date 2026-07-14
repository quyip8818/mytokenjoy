import { DEFAULT_BILLING_CURRENCY, formatCurrencyAmount } from '@/lib/currency-format'

export {
  DEFAULT_BILLING_CURRENCY,
  currencySymbol,
  formatCurrencyAmount,
} from '@/lib/currency-format'

export const DEFAULT_POINTS_PER_UNIT = 1000

export function createBillingExchange(
  pointsPerUnit: number = DEFAULT_POINTS_PER_UNIT,
  billingCurrency: string = DEFAULT_BILLING_CURRENCY,
) {
  const ppu = pointsPerUnit > 0 ? pointsPerUnit : DEFAULT_POINTS_PER_UNIT
  const currency = billingCurrency || DEFAULT_BILLING_CURRENCY
  const pointsToDisplayFn = (points: number) => points / ppu
  const displayToPointsFn = (display: number) => {
    const rounded = Math.round(display * 100) / 100
    return Math.round(rounded * ppu)
  }
  return {
    pointsPerUnit: ppu,
    billingCurrency: currency,
    pointsToDisplay: pointsToDisplayFn,
    displayToPoints: displayToPointsFn,
    formatDisplayCurrency: (points: number) =>
      formatCurrencyAmount(pointsToDisplayFn(points), currency),
    formatMoney: (amount: number) => formatCurrencyAmount(amount, currency),
  }
}

export type BillingExchange = ReturnType<typeof createBillingExchange>

let active = createBillingExchange()

export function setActiveBillingExchange(exchange: BillingExchange): void {
  active = exchange
}

export function getActiveBillingExchange(): BillingExchange {
  return active
}

export function pointsToDisplay(points: number): number {
  return active.pointsToDisplay(points)
}

export function displayToPoints(display: number): number {
  return active.displayToPoints(display)
}

/** Format point amounts (budget / key limits) using active company exchange. */
export function formatDisplayCurrency(points: number): string {
  return active.formatDisplayCurrency(points)
}

/** Format amounts already in display currency using active company currency. */
export function formatMoney(amount: number, currency?: string): string {
  return formatCurrencyAmount(amount, currency ?? active.billingCurrency)
}

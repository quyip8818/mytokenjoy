import { DEFAULT_BILLING_CURRENCY, formatCurrencyAmount } from '@/lib/currency-format'

export {
  DEFAULT_BILLING_CURRENCY,
  currencySymbol,
  formatCurrencyAmount,
} from '@/lib/currency-format'

// 1 CNY = 500000 quota, aligned with NewAPI QuotaPerUnit.
export const DEFAULT_QUOTA_PER_UNIT = 500000

export function createBillingExchange(
  quotaPerUnit: number = DEFAULT_QUOTA_PER_UNIT,
  billingCurrency: string = DEFAULT_BILLING_CURRENCY,
) {
  const qpu = quotaPerUnit > 0 ? quotaPerUnit : DEFAULT_QUOTA_PER_UNIT
  const currency = billingCurrency || DEFAULT_BILLING_CURRENCY
  const quotaToDisplayFn = (quota: number) => (qpu > 0 ? quota / qpu : 0)
  const displayToQuotaFn = (display: number) => Math.round(display * qpu)
  return {
    quotaPerUnit: qpu,
    billingCurrency: currency,
    quotaToDisplay: quotaToDisplayFn,
    displayToQuota: displayToQuotaFn,
    formatDisplayCurrency: (quota: number) =>
      formatCurrencyAmount(quotaToDisplayFn(quota), currency),
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

/** Convert quota to display amount (e.g. CNY). */
export function quotaToDisplay(quota: number): number {
  return active.quotaToDisplay(quota)
}

/** Convert display amount to quota. */
export function displayToQuota(display: number): number {
  return active.displayToQuota(display)
}

/** Format quota amounts (budget / key limits) using active company exchange. */
export function formatDisplayCurrency(quota: number): string {
  return active.formatDisplayCurrency(quota)
}

/** Format amounts already in display currency using active company currency. */
export function formatMoney(amount: number, currency?: string): string {
  return formatCurrencyAmount(amount, currency ?? active.billingCurrency)
}

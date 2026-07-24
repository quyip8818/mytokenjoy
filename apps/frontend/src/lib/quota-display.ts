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
  const quotaToMoneyFn = (quota: number) => (qpu > 0 ? quota / qpu : 0)
  const moneyToQuotaFn = (money: number) => Math.round(money * qpu)
  return {
    quotaPerUnit: qpu,
    billingCurrency: currency,
    quotaToMoney: quotaToMoneyFn,
    moneyToQuota: moneyToQuotaFn,
    formatQuotaAsMoney: (quota: number) =>
      formatCurrencyAmount(quotaToMoneyFn(quota), currency),
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

/** Convert quota to money amount (e.g. CNY). */
export function quotaToMoney(quota: number): number {
  return active.quotaToMoney(quota)
}

/** Convert money amount to quota. */
export function moneyToQuota(money: number): number {
  return active.moneyToQuota(money)
}

/** Format quota as money string using active company exchange (wallet / NewAPI quota). */
export function formatQuotaAsMoney(quota: number): string {
  return active.formatQuotaAsMoney(quota)
}

/** Format budget/spend amounts already in company billing currency. */
export function formatMoney(amount: number, currency?: string): string {
  return formatCurrencyAmount(amount, currency ?? active.billingCurrency)
}

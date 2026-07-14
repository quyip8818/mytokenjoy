export const DEFAULT_BILLING_CURRENCY = 'CNY'

/** Common billing currency symbols; unknown codes prefix with "CODE ". */
export function currencySymbol(currency: string): string {
  switch (currency.toUpperCase()) {
    case 'CNY':
    case 'RMB':
      return '¥'
    case 'USD':
      return '$'
    case 'EUR':
      return '€'
    default:
      return currency ? `${currency} ` : '¥'
  }
}

/** Format an amount already in display currency (CallLog / wallet / Spend). */
export function formatCurrencyAmount(
  amount: number,
  currency: string = DEFAULT_BILLING_CURRENCY,
): string {
  return `${currencySymbol(currency)}${Number(amount).toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`
}

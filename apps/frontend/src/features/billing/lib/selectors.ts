import type { WalletView, WalletCurrencyView } from '@/api/billing'

export function primaryWalletBalance(
  wallet: WalletView | undefined,
): WalletCurrencyView | undefined {
  if (!wallet?.balances) return undefined
  return (
    wallet.balances.find((entry) => entry.currency === wallet.billingCurrency) ?? wallet.balances[0]
  )
}

export function walletBillingCurrency(wallet: WalletView | undefined): string {
  return wallet?.billingCurrency ?? 'CNY'
}

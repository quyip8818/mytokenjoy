import { type ReactNode } from 'react'
import { setActiveBillingExchange, type BillingExchange } from '@/lib/points'
import { BillingExchangeContext } from './billing-exchange-context'

/** Syncs active module exchange during render (before children) and provides DI via context. */
export function BillingExchangeProvider({
  exchange,
  children,
}: {
  exchange: BillingExchange
  children: ReactNode
}) {
  setActiveBillingExchange(exchange)
  return (
    <BillingExchangeContext.Provider value={exchange}>{children}</BillingExchangeContext.Provider>
  )
}

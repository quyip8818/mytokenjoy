import { createContext, useContext, type ReactNode } from 'react'
import {
  getActiveBillingExchange,
  setActiveBillingExchange,
  type BillingExchange,
} from '@/lib/points'

const BillingExchangeContext = createContext<BillingExchange | null>(null)

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

export function useBillingExchange(): BillingExchange {
  return useContext(BillingExchangeContext) ?? getActiveBillingExchange()
}

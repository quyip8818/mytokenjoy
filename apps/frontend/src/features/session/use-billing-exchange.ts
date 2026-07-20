import { useContext } from 'react'
import { getActiveBillingExchange, type BillingExchange } from '@/lib/quota-display'
import { BillingExchangeContext } from './billing-exchange-context'

export function useBillingExchange(): BillingExchange {
  return useContext(BillingExchangeContext) ?? getActiveBillingExchange()
}

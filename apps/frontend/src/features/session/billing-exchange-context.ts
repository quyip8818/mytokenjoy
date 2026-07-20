import { createContext } from 'react'
import type { BillingExchange } from '@/lib/quota-display'

export const BillingExchangeContext = createContext<BillingExchange | null>(null)

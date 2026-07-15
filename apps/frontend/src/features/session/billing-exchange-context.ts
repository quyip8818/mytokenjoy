import { createContext } from 'react'
import type { BillingExchange } from '@/lib/points'

export const BillingExchangeContext = createContext<BillingExchange | null>(null)

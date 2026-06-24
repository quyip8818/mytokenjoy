import { createContext } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import type { PageContextState } from './page-context-store'

export const PageContextStoreContext = createContext<StoreApi<PageContextState> | null>(null)

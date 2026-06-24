import { createContext } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import type { DemoGuideStoreState } from './store'

export const DemoGuideStoreContext = createContext<StoreApi<DemoGuideStoreState> | null>(null)

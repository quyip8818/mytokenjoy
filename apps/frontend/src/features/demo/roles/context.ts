import { createContext } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import type { DemoRoleStoreState } from './store'

export const DemoRoleStoreContext = createContext<StoreApi<DemoRoleStoreState> | null>(null)

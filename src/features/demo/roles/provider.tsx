import { useMemo, type ReactNode } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import { DemoRoleStoreContext } from './context'
import { defaultDemoRoleStore, type DemoRoleStoreState } from './store'

interface DemoRoleProviderProps {
  children: ReactNode
  store?: StoreApi<DemoRoleStoreState>
}

export function DemoRoleProvider({ children, store }: DemoRoleProviderProps) {
  const storeInstance = useMemo(() => store ?? defaultDemoRoleStore, [store])
  return (
    <DemoRoleStoreContext.Provider value={storeInstance}>{children}</DemoRoleStoreContext.Provider>
  )
}

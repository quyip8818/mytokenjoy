import { useMemo, type ReactNode } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import { DemoGuideStoreContext } from './context'
import { createDemoGuideStore, type DemoGuideStoreState } from './store'

interface DemoGuideProviderProps {
  children: ReactNode
  store?: StoreApi<DemoGuideStoreState>
}

export function DemoGuideProvider({ children, store }: DemoGuideProviderProps) {
  const storeInstance = useMemo(() => store ?? createDemoGuideStore(), [store])
  return (
    <DemoGuideStoreContext.Provider value={storeInstance}>
      {children}
    </DemoGuideStoreContext.Provider>
  )
}

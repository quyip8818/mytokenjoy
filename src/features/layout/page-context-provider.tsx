import { useMemo, type ReactNode } from 'react'
import type { StoreApi } from 'zustand/vanilla'
import { defaultPageContextStore } from './page-context-store'
import type { PageContextState } from './page-context-store'
import { PageContextStoreContext } from './page-context-store-context'

interface PageContextProviderProps {
  children: ReactNode
  store?: StoreApi<PageContextState>
}

export function PageContextProvider({ children, store }: PageContextProviderProps) {
  const storeInstance = useMemo(() => store ?? defaultPageContextStore, [store])
  return (
    <PageContextStoreContext.Provider value={storeInstance}>
      {children}
    </PageContextStoreContext.Provider>
  )
}

import { useContext } from 'react'
import { useStore } from 'zustand'
import { PageContextStoreContext } from './page-context-store-context'

export function usePageContext() {
  const store = useContext(PageContextStoreContext)
  if (!store) {
    throw new Error('usePageContext must be used within PageContextProvider')
  }
  const subtitle = useStore(store, (s) => s.subtitle)
  const setSubtitle = useStore(store, (s) => s.setSubtitle)
  return { subtitle, setSubtitle }
}

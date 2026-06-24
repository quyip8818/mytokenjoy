import { createStore, type StoreApi } from 'zustand/vanilla'

export interface PageContextState {
  subtitle: string | null
  setSubtitle: (subtitle: string | null) => void
}

export function createPageContextStore(): StoreApi<PageContextState> {
  return createStore<PageContextState>((set) => ({
    subtitle: null,
    setSubtitle: (subtitle) => set({ subtitle }),
  }))
}

export const defaultPageContextStore = createPageContextStore()

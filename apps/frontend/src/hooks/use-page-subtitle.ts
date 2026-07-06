import { createStore, type StoreApi } from 'zustand/vanilla'
import { useStore } from 'zustand'

export interface PageSubtitleState {
  subtitle: string | null
  setSubtitle: (subtitle: string | null) => void
}

export function createPageSubtitleStore(): StoreApi<PageSubtitleState> {
  return createStore<PageSubtitleState>((set) => ({
    subtitle: null,
    setSubtitle: (subtitle) => set({ subtitle }),
  }))
}

export const defaultPageSubtitleStore = createPageSubtitleStore()

export function usePageSubtitle(store: StoreApi<PageSubtitleState> = defaultPageSubtitleStore) {
  const subtitle = useStore(store, (s) => s.subtitle)
  const setSubtitle = useStore(store, (s) => s.setSubtitle)
  return { subtitle, setSubtitle }
}

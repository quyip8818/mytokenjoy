import { useContext } from 'react'
import { useStore } from 'zustand'
import type { StoreApi } from 'zustand/vanilla'
import { cn } from '@/lib/utils'
import { DemoGuideStoreContext } from './context'
import type { DemoGuideStoreState } from './store'
import { DEMO_CTA_IDS, type DemoCtaKey } from './constants'

function useDemoGuideStoreApi(): StoreApi<DemoGuideStoreState> {
  const store = useContext(DemoGuideStoreContext)
  if (!store) throw new Error('useDemoGuide must be used within DemoGuideProvider')
  return store
}

export function useDemoGuide() {
  const store = useDemoGuideStoreApi()
  return useStore(store)
}

export function useDemoGuideHighlight(ctaId: string): string {
  const store = useDemoGuideStoreApi()
  const highlightCtaId = useStore(store, (s) => s.highlightCtaId)
  if (highlightCtaId !== ctaId) return ''
  return cn('ring-2 ring-blue-400/70 ring-offset-2 shadow-[0_0_12px_rgba(37,99,235,0.35)]')
}

export function useDemoCta(key: DemoCtaKey) {
  const id = DEMO_CTA_IDS[key]
  const className = useDemoGuideHighlight(id)
  return { id, className }
}

import { useContext, useCallback, type ReactNode } from 'react'
import { useStore } from 'zustand'
import { cn } from '@/lib/utils'
import {
  CtaHighlightContext,
  type CtaHighlightKey,
  type CtaHighlightResolver,
} from '@/hooks/use-cta-highlight'
import { DEMO_CTA_IDS } from './constants'
import { DemoGuideStoreContext } from './context'

export function DemoCtaHighlightProvider({ children }: { children: ReactNode }) {
  const store = useContext(DemoGuideStoreContext)
  if (!store) {
    throw new Error('DemoCtaHighlightProvider must be used within DemoGuideProvider')
  }

  const highlightCtaId = useStore(store, (state) => state.highlightCtaId)

  const resolver = useCallback<CtaHighlightResolver>(
    (key: CtaHighlightKey) => {
      const id = DEMO_CTA_IDS[key]
      const className =
        highlightCtaId === id
          ? cn('ring-2 ring-blue-400/70 ring-offset-2 shadow-[0_0_12px_rgba(37,99,235,0.35)]')
          : ''
      return { id, className }
    },
    [highlightCtaId],
  )

  return <CtaHighlightContext.Provider value={resolver}>{children}</CtaHighlightContext.Provider>
}

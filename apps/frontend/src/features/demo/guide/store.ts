import { createStore, type StoreApi } from 'zustand/vanilla'
import { DEMO_GUIDE_STORAGE_KEY, DEMO_GUIDE_STEPS } from './constants'

function loadCompleted(): Set<string> {
  try {
    const raw = localStorage.getItem(DEMO_GUIDE_STORAGE_KEY)
    if (!raw) return new Set()
    return new Set(JSON.parse(raw) as string[])
  } catch {
    return new Set()
  }
}

function persistCompleted(completed: Set<string>) {
  localStorage.setItem(DEMO_GUIDE_STORAGE_KEY, JSON.stringify([...completed]))
}

export interface DemoGuideStoreState {
  open: boolean
  highlightCtaId: string | null
  completed: Set<string>
  setOpen: (open: boolean) => void
  setHighlightCtaId: (ctaId: string | null) => void
  markComplete: (stepId: string) => void
  resetProgress: () => void
  isStepComplete: (stepId: string) => boolean
}

export function createDemoGuideStore(): StoreApi<DemoGuideStoreState> {
  return createStore<DemoGuideStoreState>((set, get) => ({
    open: false,
    highlightCtaId: null,
    completed: loadCompleted(),
    setOpen: (open) => set({ open, highlightCtaId: open ? get().highlightCtaId : null }),
    setHighlightCtaId: (highlightCtaId) => set({ highlightCtaId }),
    markComplete: (stepId) => {
      const next = new Set(get().completed)
      next.add(stepId)
      persistCompleted(next)
      set({ completed: next })
    },
    resetProgress: () => {
      localStorage.removeItem(DEMO_GUIDE_STORAGE_KEY)
      set({ completed: new Set(), highlightCtaId: null })
    },
    isStepComplete: (stepId) => get().completed.has(stepId),
  }))
}

export function getDemoGuideStepCount(): number {
  return DEMO_GUIDE_STEPS.length
}

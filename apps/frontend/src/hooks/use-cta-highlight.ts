import { createContext, useContext } from 'react'

export type CtaHighlightKey =
  | 'CREDENTIAL'
  | 'IMPORT'
  | 'BUDGET'
  | 'MODEL'
  | 'OVERRUN'
  | 'CREATE_KEY'
  | 'APPLY_BUDGET'

export interface CtaHighlightResult {
  id: string
  className: string
}

export type CtaHighlightResolver = (key: CtaHighlightKey) => CtaHighlightResult

const noopResolver: CtaHighlightResolver = () => ({ id: '', className: '' })

export const CtaHighlightContext = createContext<CtaHighlightResolver | null>(null)

export function useCtaHighlight(key: CtaHighlightKey): CtaHighlightResult {
  const resolver = useContext(CtaHighlightContext)
  return (resolver ?? noopResolver)(key)
}

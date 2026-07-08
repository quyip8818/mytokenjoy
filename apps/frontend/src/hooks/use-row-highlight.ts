import { useCallback, useState } from 'react'

const HIGHLIGHT_DURATION_MS = 2000

export function useRowHighlight() {
  const [highlightId, setHighlightId] = useState<string | number | null>(null)

  const flashRow = useCallback((id: string | number) => {
    setHighlightId(id)
    const timer = window.setTimeout(() => setHighlightId(null), HIGHLIGHT_DURATION_MS)
    return () => window.clearTimeout(timer)
  }, [])

  const rowClass = useCallback(
    (id: string | number, base = 'border-border/40 hover:bg-blue-50/30') =>
      `${base} transition-colors ${highlightId === id ? 'bg-blue-50/60' : ''}`,
    [highlightId],
  )

  return { highlightId, flashRow, rowClass }
}

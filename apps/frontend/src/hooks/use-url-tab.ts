import { useCallback } from 'react'
import { useSearchParams } from 'react-router'

/**
 * Sync a tab value with URL search params for deep link support.
 *
 * @param validTabs - Array of valid tab values
 * @param defaultTab - Default tab (won't appear in URL to keep it clean)
 * @param paramName - URL query param name (default: "tab")
 *
 * @example
 * const [tab, setTab] = useUrlTab(['account', 'security', 'notifications'], 'account')
 */
export function useUrlTab<T extends string>(
  validTabs: readonly T[],
  defaultTab: T,
  paramName = 'tab',
): [T, (tab: T) => void] {
  const [searchParams, setSearchParams] = useSearchParams()
  const raw = searchParams.get(paramName)
  const activeTab = raw && validTabs.includes(raw as T) ? (raw as T) : defaultTab

  const setTab = useCallback(
    (tab: T) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          if (tab === defaultTab) {
            next.delete(paramName)
          } else {
            next.set(paramName, tab)
          }
          return next
        },
        { replace: true },
      )
    },
    [setSearchParams, defaultTab, paramName],
  )

  return [activeTab, setTab]
}

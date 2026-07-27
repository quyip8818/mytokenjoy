import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { SIDEBAR_COLLAPSED_STORAGE_KEY } from './sidebar-layout-constants'
import { SidebarLayoutContext } from './sidebar-layout-context'

function readCollapsed(): boolean {
  try {
    return localStorage.getItem(SIDEBAR_COLLAPSED_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function persistCollapsed(collapsed: boolean) {
  try {
    localStorage.setItem(SIDEBAR_COLLAPSED_STORAGE_KEY, String(collapsed))
  } catch {
    // ignore storage failures
  }
}

export function SidebarLayoutProvider({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(readCollapsed)

  useEffect(() => {
    persistCollapsed(collapsed)
  }, [collapsed])

  const toggleCollapsed = useCallback(() => {
    setCollapsed((prev) => !prev)
  }, [])

  const value = useMemo(
    () => ({
      collapsed,
      toggleCollapsed,
    }),
    [collapsed, toggleCollapsed],
  )

  return <SidebarLayoutContext.Provider value={value}>{children}</SidebarLayoutContext.Provider>
}

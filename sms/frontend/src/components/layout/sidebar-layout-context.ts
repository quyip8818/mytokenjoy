import { createContext } from 'react'

export interface SidebarLayoutContextValue {
  collapsed: boolean
  toggleCollapsed: () => void
}

export const SidebarLayoutContext = createContext<SidebarLayoutContextValue | null>(null)

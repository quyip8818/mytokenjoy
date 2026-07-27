import { useContext } from 'react'
import { SidebarLayoutContext } from './sidebar-layout-context'

export function useSidebarLayout() {
  const context = useContext(SidebarLayoutContext)
  if (!context) {
    throw new Error('useSidebarLayout must be used within SidebarLayoutProvider')
  }
  return context
}

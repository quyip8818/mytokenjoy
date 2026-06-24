import { useContext } from 'react'
import { useStore } from 'zustand'
import type { StoreApi } from 'zustand/vanilla'
import { DemoRoleStoreContext } from './context'
import type { DemoRoleStoreState } from './store'

function useDemoRoleStoreApi(): StoreApi<DemoRoleStoreState> {
  const store = useContext(DemoRoleStoreContext)
  if (!store) throw new Error('useDemoRole must be used within DemoRoleProvider')
  return store
}

export function useDemoRole() {
  const store = useDemoRoleStoreApi()
  return useStore(store)
}

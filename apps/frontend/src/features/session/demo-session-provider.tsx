import { useContext, useMemo, type ReactNode } from 'react'
import { useStore } from 'zustand'
import type { StoreApi } from 'zustand/vanilla'
import { DemoRoleProvider } from '@/features/demo/roles/provider'
import { DemoRoleStoreContext } from '@/features/demo/roles/context'
import type { DemoRoleStoreState } from '@/features/demo/roles/store'
import { SessionReactContext } from './context'
import { SessionGate } from './session-gate'
import type { AppSession } from './types'

interface DemoSessionProviderProps {
  children: ReactNode
  store?: StoreApi<DemoRoleStoreState>
}

function DemoSessionBridge({ children }: { children: ReactNode }) {
  const store = useContext(DemoRoleStoreContext)
  if (!store) {
    throw new Error('DemoSessionBridge must be used within DemoRoleProvider')
  }

  const memberId = useStore(store, (s) => s.memberId)
  const member = useStore(store, (s) => s.member)
  const permissions = useStore(store, (s) => s.permissions)
  const readOnly = useStore(store, (s) => s.readOnly)
  const loading = useStore(store, (s) => s.loading)
  const sessionError = useStore(store, (s) => s.sessionError)
  const refreshSession = useStore(store, (s) => s.refreshSession)

  const session = useMemo<AppSession>(
    () => ({
      memberId,
      member,
      permissions,
      readOnly,
      loading,
      sessionError,
      refreshSession,
    }),
    [memberId, member, permissions, readOnly, loading, sessionError, refreshSession],
  )

  return (
    <SessionReactContext.Provider value={session}>
      <SessionGate>{children}</SessionGate>
    </SessionReactContext.Provider>
  )
}

export function DemoSessionProvider({ children, store }: DemoSessionProviderProps) {
  return (
    <DemoRoleProvider store={store}>
      <DemoSessionBridge>{children}</DemoSessionBridge>
    </DemoRoleProvider>
  )
}

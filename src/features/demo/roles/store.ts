import { createStore, type StoreApi } from 'zustand/vanilla'
import { DEMO_ROLE_PROFILES, type DemoRole } from './constants'

export interface DemoRoleStoreState {
  role: DemoRole
  memberId: string
  displayName: string
  initials: string
  setRole: (role: DemoRole) => void
}

export function createDemoRoleStore(initialRole: DemoRole = 'admin'): StoreApi<DemoRoleStoreState> {
  const profile = DEMO_ROLE_PROFILES[initialRole]
  return createStore<DemoRoleStoreState>((set) => ({
    role: initialRole,
    memberId: profile.memberId,
    displayName: profile.displayName,
    initials: profile.initials,
    setRole: (role) => {
      const next = DEMO_ROLE_PROFILES[role]
      set({
        role,
        memberId: next.memberId,
        displayName: next.displayName,
        initials: next.initials,
      })
    },
  }))
}

export const defaultDemoRoleStore = createDemoRoleStore()

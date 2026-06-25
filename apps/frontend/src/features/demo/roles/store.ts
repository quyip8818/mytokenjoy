import { createStore, type StoreApi } from 'zustand/vanilla'
import { toast } from 'sonner'
import { sessionApi } from '@/api/session'
import { setDemoMemberIdProvider } from '@/api/client'
import type { Member } from '@/api/types'
import {
  DEFAULT_DEMO_MEMBER_ID,
  DEMO_SWITCHABLE_MEMBERS,
  getMemberDisplay,
  getSwitchableMember,
} from './constants'

export interface DemoRoleStoreState {
  memberId: string
  member: Member | null
  permissions: string[]
  readOnly: boolean
  roles: string[]
  displayName: string
  initials: string
  loading: boolean
  setMemberId: (memberId: string) => Promise<void>
  refreshSession: () => Promise<void>
}

function profileFromMember(member: Member) {
  const switchable = getSwitchableMember(member.id)
  if (switchable) {
    return {
      displayName: switchable.displayName,
      initials: switchable.initials,
    }
  }
  return getMemberDisplay(member)
}

export function createDemoRoleStore(
  initialMemberId: string = DEFAULT_DEMO_MEMBER_ID,
): StoreApi<DemoRoleStoreState> {
  const fallback = DEMO_SWITCHABLE_MEMBERS.find((m) => m.id === initialMemberId)
  const store = createStore<DemoRoleStoreState>((set, get) => ({
    memberId: initialMemberId,
    member: null,
    permissions: [],
    readOnly: false,
    roles: [],
    displayName: fallback?.displayName ?? '用户',
    initials: fallback?.initials ?? '?',
    loading: true,
    setMemberId: async (memberId: string) => {
      set({ memberId, loading: true })
      await get().refreshSession()
    },
    refreshSession: async () => {
      const { memberId } = get()
      try {
        const session = await sessionApi.get(memberId)
        const profile = profileFromMember(session.member)
        set({
          member: session.member,
          permissions: session.permissions,
          readOnly: session.readOnly,
          roles: session.member.roles,
          displayName: profile.displayName,
          initials: profile.initials,
          loading: false,
        })
      } catch {
        set({
          member: null,
          permissions: [],
          readOnly: false,
          roles: [],
          loading: false,
        })
        toast.error('Failed to load session')
      }
    },
  }))

  setDemoMemberIdProvider(() => store.getState().memberId)
  void store.getState().refreshSession()

  return store
}

export const defaultDemoRoleStore = createDemoRoleStore()

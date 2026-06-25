import { createStore, type StoreApi } from 'zustand/vanilla'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
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
  sessionError: Error | null
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

function resolveInitialMemberId(memberId: string): string {
  if (DEMO_SWITCHABLE_MEMBERS.some((m) => m.id === memberId)) {
    return memberId
  }
  if (import.meta.env.DEV) {
    console.warn(
      `[Demo] Unknown member id "${memberId}", using default "${DEFAULT_DEMO_MEMBER_ID}"`,
    )
  }
  return DEFAULT_DEMO_MEMBER_ID
}

export function createDemoRoleStore(
  initialMemberId: string = DEFAULT_DEMO_MEMBER_ID,
  apis: Pick<AppApis, 'sessionApi'> = defaultApis,
): StoreApi<DemoRoleStoreState> {
  const resolvedMemberId = resolveInitialMemberId(initialMemberId)
  const fallback = DEMO_SWITCHABLE_MEMBERS.find((m) => m.id === resolvedMemberId)
  const store = createStore<DemoRoleStoreState>((set, get) => ({
    memberId: resolvedMemberId,
    member: null,
    permissions: [],
    readOnly: false,
    roles: [],
    displayName: fallback?.displayName ?? '用户',
    initials: fallback?.initials ?? '?',
    loading: true,
    sessionError: null,
    setMemberId: async (memberId: string) => {
      set({ memberId, loading: true, sessionError: null })
      await get().refreshSession()
    },
    refreshSession: async () => {
      const { memberId } = get()
      try {
        const session = await apis.sessionApi.get(memberId)
        const profile = profileFromMember(session.member)
        set({
          member: session.member,
          permissions: session.permissions,
          readOnly: session.readOnly,
          roles: session.member.roles,
          displayName: profile.displayName,
          initials: profile.initials,
          loading: false,
          sessionError: null,
        })
      } catch (error) {
        const sessionError = error instanceof Error ? error : new Error(String(error))
        set({
          loading: false,
          sessionError,
        })
      }
    },
  }))

  setDemoMemberIdProvider(() => store.getState().memberId)
  void store.getState().refreshSession()

  return store
}

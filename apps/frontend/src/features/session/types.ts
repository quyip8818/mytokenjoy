import type { Member } from '@/api/types'

export interface AppSession {
  companyId: number
  authzRevision: number
  memberId: string
  member: Member | null
  permissions: string[]
  readOnly: boolean
  loading: boolean
  sessionError: Error | null
  refreshSession: () => Promise<void>
}

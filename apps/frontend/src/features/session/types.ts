import type { Member } from '@/api/types'
import type { CompanyType } from '@/api/types/common'

export type { CompanyType } from '@/api/types/common'

export interface AppSession {
  companyId: string
  companyName: string
  companyType: CompanyType
  authzRevision: number
  memberId: string
  member: Member | null
  userName: string
  permissions: string[]
  readOnly: boolean
  billingCurrency: string
  quotaPerUnit: number
  loading: boolean
  sessionError: Error | null
  refreshSession: () => Promise<void>
}

import type { Member } from '@/api/types'
import type { CompanyType } from '@/api/types/common'

export type { CompanyType } from '@/api/types/common'

export interface AppSession {
  companyId: number
  companyType: CompanyType
  authzRevision: number
  memberId: string
  member: Member | null
  permissions: string[]
  readOnly: boolean
  billingCurrency: string
  pointsPerUnit: number
  loading: boolean
  sessionError: Error | null
  refreshSession: () => Promise<void>
}

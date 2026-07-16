import type { Member } from '@/api/types'

export type CompanyType = 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'

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

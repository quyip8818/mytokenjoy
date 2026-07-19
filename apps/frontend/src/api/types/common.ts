import type { Member } from './org'

export type CompanyType = 'standard' | 'trial' | 'demo' | 'selfhosted' | 'testing'

export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface MemberPaginated extends Paginated<Member> {
  pendingCount: number
}

export interface SessionContext {
  companyId: string
  companyType: CompanyType
  authzRevision: number
  member: Member
  permissions: string[]
  readOnly: boolean
  billingCurrency: string
  quotaPerUnit: number
}

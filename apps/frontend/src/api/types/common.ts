import type { Member } from './org'

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
  companyId: number
  authzRevision: number
  member: Member
  permissions: string[]
  readOnly: boolean
  billingCurrency: string
  pointsPerUnit: number
}

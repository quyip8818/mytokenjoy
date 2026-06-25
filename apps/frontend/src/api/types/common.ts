import type { Member } from './org'

export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface SessionContext {
  member: Member
  permissions: string[]
  readOnly: boolean
}

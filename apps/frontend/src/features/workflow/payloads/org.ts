import type { Member } from '@/api/types'

export interface OrgWorkflowPayloads {
  'member-search': {
    excludeIds?: string[]
    multi?: boolean
    onConfirm?: (members: Member[]) => void
  }
}

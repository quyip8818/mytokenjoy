import type { Permission } from '@/api/types'

export interface SharedWorkflowPayloads {
  'quota-check': {
    reservedPool?: number
    requested?: number
  }
  'permission-picker': {
    permissions?: Permission[]
    selected?: string[]
    onConfirm?: (perms: string[]) => void
  }
}

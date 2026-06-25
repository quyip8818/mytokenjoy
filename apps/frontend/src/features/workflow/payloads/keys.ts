import type { KeyApproval, PlatformKey } from '@/api/types'

export interface KeysWorkflowPayloads {
  'key-create': {
    adminCreate?: boolean
    targetMemberId?: string
    budgetGroupId?: string
    budgetGroupName?: string
    onSuccess?: (id?: string) => void
  }
  'key-edit': {
    key?: PlatformKey
    adminCreate?: boolean
    targetMemberId?: string
    budgetGroupId?: string
    budgetGroupName?: string
    onSuccess?: (id?: string) => void
  }
  'key-rotate-confirm': {
    key: PlatformKey
    onRotate?: (key: PlatformKey) => Promise<{ fullKey?: string; keyPrefix: string }>
    onDone?: () => void
  }
  'key-reveal': {
    fullKey?: string
    onDone?: () => void
  }
  'approval-submit': {
    defaultType?: 'key' | 'quota'
    onSuccess?: () => void
  }
  'approval-review': {
    approval: KeyApproval
    onSuccess?: () => void
  }
  'provider-key-form': {
    onSuccess?: () => void
  }
  'reject-reason': {
    approvalId: string
    onSuccess?: () => void
  }
}

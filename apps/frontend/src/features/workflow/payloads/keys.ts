import type { ApprovalRequest, PlatformKey, PlatformKeyScope } from '@/api/types'

export interface KeysWorkflowPayloads {
  'key-create': {
    adminCreate?: boolean
    scope: PlatformKeyScope
    targetMemberId?: string
    projectId?: string
    projectName?: string
    initialName?: string
    initialBudget?: string
    onSuccess?: (id?: string) => void
  }
  'key-edit': {
    key?: PlatformKey
    adminCreate?: boolean
    targetMemberId?: string
    projectId?: string
    projectName?: string
    initialName?: string
    initialBudget?: string
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
    defaultType?: 'key' | 'member_budget'
    onSuccess?: () => void
  }
  'approval-review': {
    approval: ApprovalRequest
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

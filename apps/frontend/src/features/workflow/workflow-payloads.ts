import type {
  BatchImportRow,
  BudgetGroup,
  BudgetNode,
  Department,
  KeyApproval,
  Member,
  Permission,
  Platform,
  PlatformKey,
  Role,
  RoutingRule,
} from '@/api/types'

export interface MemberFormData {
  name: string
  phone: string
  email: string
  departmentId: string
}

export interface WorkflowPayloadMap {
  'credential-form': {
    connected?: boolean
    currentPlatform?: Platform | null
    onSuccess?: () => void
  }
  'sync-config': {
    onTriggerSync?: () => void
    triggeringSync?: boolean
    onSuccess?: () => void
  }
  'member-form': {
    member?: Member | null
    departments?: Department[]
    defaultDeptId?: string
    onSubmit?: (data: MemberFormData) => Promise<void>
  }
  'member-invite': {
    onSubmit?: (value: string) => Promise<void>
  }
  'member-import': {
    defaultDeptName?: string
    onSuccess?: () => void
  }
  'dept-form': {
    department?: Department | null
    parentId?: string
    parentName?: string
    onSuccess?: () => void
  }
  'budget-node-edit': {
    node: BudgetNode
    parent?: BudgetNode | null
    onSuccess?: () => void
  }
  'budget-group-form': {
    group?: BudgetGroup
    tree?: BudgetNode[]
    onSuccess?: (id?: string) => void
  }
  'overrun-policy': {
    onSuccess?: () => void
  }
  'model-create': {
    onSuccess?: (id?: string) => void
  }
  'whitelist-config': {
    rule: RoutingRule
    onSuccess?: () => void
  }
  'key-create': {
    adminCreate?: boolean
    targetMemberId?: string
    onSuccess?: (id?: string) => void
  }
  'key-edit': {
    key?: PlatformKey
    adminCreate?: boolean
    targetMemberId?: string
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
  'role-form': {
    role?: Role | null
    permissions?: Permission[]
    onSubmit?: (data: { name: string; permissions: string[] }) => Promise<void>
  }
  'role-add-member': {
    roleId: string
    roleName?: string
    existingMemberIds?: string[]
    onSuccess?: () => void
  }
  'provider-key-form': {
    onSuccess?: () => void
  }
  'pick-dept': {
    selectedId?: string
    onConfirm?: (deptId: string) => void
  }
  'model-picker': {
    selectedModels?: string[]
    parentWhitelist?: string[]
    onConfirm?: (models: string[]) => void
  }
  'import-preview': {
    rows?: BatchImportRow[]
    onSuccess?: () => void
  }
  'quota-check': {
    reservedPool?: number
    requested?: number
  }
  'reject-reason': {
    approvalId: string
    onSuccess?: () => void
  }
  'budget-impact-preview': {
    nodeId: string
    nodeName: string
    before: { budget: number; reservedPool: number }
    after: { budget: number; reservedPool: number }
    onSuccess?: () => void
  }
  'permission-picker': {
    permissions?: Permission[]
    selected?: string[]
    onConfirm?: (perms: string[]) => void
  }
  'member-search': {
    excludeIds?: string[]
    multi?: boolean
    onConfirm?: (members: Member[]) => void
  }
  'pick-members': {
    departmentId: string
    selectedIds?: string[]
    onConfirm?: (memberIds: string[], members: Member[]) => void
  }
}

export type WorkflowPayload<T extends keyof WorkflowPayloadMap = keyof WorkflowPayloadMap> =
  WorkflowPayloadMap[T]

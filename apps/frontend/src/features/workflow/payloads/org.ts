import type { Department, Member, Permission, Role } from '@/api/types'
import type { BatchImportRow } from '@/api/types'

export interface MemberFormData {
  name: string
  phone: string
  email: string
  departmentId: string
}

export interface OrgWorkflowPayloads {
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
  'import-preview': {
    rows?: BatchImportRow[]
    onSuccess?: () => void
  }
  'pick-dept': {
    selectedId?: string
    onConfirm?: (deptId: string) => void
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

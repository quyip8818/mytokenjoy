export type Platform = 'feishu' | 'dingtalk' | 'wecom'

export interface FeishuCredential {
  platform: 'feishu'
  appId: string
  appSecret: string
}

export interface DingtalkCredential {
  platform: 'dingtalk'
  corpId: string
  appKey: string
  appSecret: string
}

export interface WecomCredential {
  platform: 'wecom'
  corpId: string
  secret: string
  agentId: string
}

export type Credential = FeishuCredential | DingtalkCredential | WecomCredential

export interface DataSourceStatus {
  platform: Platform | null
  connected: boolean
  lastImport: string | null
  lastImportResult: ImportResult | null
}

export interface ImportResult {
  successMembers: number
  successDepartments: number
  failures: ImportFailure[]
}

export interface ImportFailure {
  id: string
  name: string
  employeeId: string
  reason: string
}

export interface SyncConfig {
  enabled: boolean
  startTime: string
  frequencyHours: 6 | 12 | 24
  deleteMemberThreshold: number
  deleteDepartmentThreshold: number
  notifyPhone: boolean
  notifyEmail: boolean
  notifyIm: boolean
}

export interface SyncLog {
  id: string
  time: string
  type: 'scheduled' | 'manual'
  result: 'success' | 'partial_failure' | 'failure'
  detail: string
}

export interface Department {
  id: string
  name: string
  parentId: string | null
  children?: Department[]
  memberCount: number
  externalId?: string
  source?: 'imported' | 'manual'
  managerId?: string
}

export type MemberStatus = 'active' | 'disabled' | 'pending'

export type MemberSource =
  | 'manual'
  | 'csv'
  | 'feishu'
  | 'dingtalk'
  | 'wecom'
  | 'invited'
  | 'imported'

export type RegistrationChannel = 'sms' | 'email' | 'admin_link'

export interface Member {
  id: string
  companyId: string
  alias: string
  avatar: string
  departmentId: string
  departmentName: string
  status: MemberStatus
  roles: string[]
  source: MemberSource
  registrationChannel?: RegistrationChannel
  externalId?: string
  employeeId?: string
  jobTitle?: string
  hireDate?: string
  // Read-only fields from users table (JOIN)
  name?: string
  phone?: string
  email?: string
}

// Request body for POST /org/members
export interface CreateMemberInput {
  user: { name: string; phone: string; email: string }
  member: {
    alias?: string
    departmentId: string
    employeeId?: string
    jobTitle?: string
    hireDate?: string
  }
}

// Request body for PUT /org/members/:id
export interface UpdateMemberInput {
  alias?: string
  departmentId?: string
  employeeId?: string
  jobTitle?: string
  hireDate?: string
  roles?: string[]
  status?: string
}

// Request body for PUT /org/members/:id/user
export interface UpdateMemberUserInput {
  name?: string
  phone?: string
  email?: string
}

export interface BatchImportRow {
  name: string
  phone: string
  email: string
  departmentName: string
}

export interface MemberBatchImportResult {
  imported: number
  failures: { row: number; reason: string }[]
}

export interface Role {
  id: string
  name: string
  type: 'preset' | 'custom'
  permissions: string[]
  memberCount: number
}

export interface Permission {
  id: string
  name: string
  group: string
}

export interface FieldMapping {
  sourceField: string
  sourceLabel: string
  targetField: string
  required: boolean
}

export interface FieldMappingConfig {
  platform: Platform
  mappings: FieldMapping[]
}

export interface MappingTestResult {
  success: boolean
  preview: Record<string, string>
  errors: string[]
}

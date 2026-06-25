// 第三方平台类型
export type Platform = 'feishu' | 'dingtalk' | 'wecom'

// 凭证配置
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

// 数据源状态
export interface DataSourceStatus {
  platform: Platform | null
  connected: boolean
  lastImport: string | null
  lastImportResult: ImportResult | null
}

// 导入结果
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

// 同步配置
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

// 同步记录
export interface SyncLog {
  id: string
  time: string
  type: 'scheduled' | 'manual'
  result: 'success' | 'partial_failure' | 'failure'
  detail: string
}

// 部门
export interface Department {
  id: string
  name: string
  parentId: string | null
  children?: Department[]
  memberCount: number
}

// 成员
export type MemberStatus = 'active' | 'inactive' | 'pending'

export interface Member {
  id: string
  name: string
  phone: string
  email: string
  departmentId: string
  departmentName: string
  status: MemberStatus
  roles: string[]
  source: 'imported' | 'manual' | 'invited'
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

// 角色
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

export interface SessionContext {
  member: Member
  permissions: string[]
  readOnly: boolean
}

// 通用分页
export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

// ========== 预算管理 ==========

export interface BudgetNode {
  id: string
  name: string
  parentId: string | null
  budget: number
  consumed: number
  reservedPool?: number
  children?: BudgetNode[]
  period: string
}

export interface OverrunPolicyConfig {
  thresholds: number[]
  notifyEmail: boolean
  notifyPhone: boolean
  notifyIm: boolean
  blockMessage: string
}

export interface ResolvedWhitelist {
  inherited: boolean
  allowedModels: string[]
  parentCount: number
}

export interface CreateModelInput {
  name: string
  displayName: string
  baseUrl: string
  apiKey: string
  inputPrice: number
  outputPrice: number
}

export interface BudgetGroup {
  id: string
  name: string
  budget: number
  consumed: number
  memberIds: string[]
  departmentIds: string[]
}

export interface AlertRule {
  id: string
  nodeId: string
  nodeName: string
  thresholds: number[] // e.g. [80, 90, 100]
  notifyRoleIds: string[]
  enabled: boolean
}

// ========== API-KEY 管理 ==========

export type ProviderType = 'openai' | 'anthropic' | 'deepseek' | 'qwen' | 'custom'
export type KeyStatus = 'active' | 'disabled' | 'expired' | 'error'

export interface ProviderKey {
  id: string
  provider: ProviderType
  name: string
  keyPrefix: string // 仅展示前缀
  status: KeyStatus
  balance: number | null
  lastUsed: string | null
  createdAt: string
  rotateEnabled: boolean
}

export interface PlatformKey {
  id: string
  name: string
  keyPrefix: string
  fullKey?: string
  memberId: string | null
  memberName: string | null
  appName: string | null
  status: KeyStatus
  quota: number
  used: number
  modelWhitelist: string[]
  createdAt: string
  expiresAt: string | null
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'
export type ApprovalType = 'key' | 'quota'

export interface KeyApproval {
  id: string
  type: ApprovalType
  applicant: string
  applicantId: string
  department: string
  reason: string
  requestedQuota: number
  requestedModels: string[]
  status: ApprovalStatus
  approver: string | null
  rejectReason?: string | null
  createdAt: string
  resolvedAt: string | null
}

export interface MemberQuotaSummary {
  totalQuota: number
  used: number
  remaining: number
  reservedPool: number
}

// ========== 模型路由 ==========

export interface ModelInfo {
  id: string
  provider: ProviderType
  name: string
  displayName: string
  inputPrice: number // 每百万 token 价格
  outputPrice: number
  maxContext: number
  enabled: boolean
  capabilities: string[] // e.g. ['chat', 'vision', 'function_calling']
}

export interface RoutingRule {
  id: string
  nodeId: string
  nodeName: string
  allowedModels: string[]
  defaultModel: string | null
  fallbackModel: string | null
  inherited: boolean
}

// ========== 数据看板 ==========

export interface CostSummary {
  totalCost: number
  monthOverMonth: number
  totalTokens: number
  totalRequests: number
  avgCostPerRequest: number
  avgCostPerMember: number
}

export type CostPeriod = 'current_month' | 'last_month' | 'last_7_days'

export interface DepartmentCost {
  departmentId: string
  departmentName: string
  cost: number
  percentage: number
  hasChildren?: boolean
}

export interface DepartmentCostMember {
  memberId: string
  memberName: string
  cost: number
  requests: number
  tokens: number
}

export interface DailyCost {
  date: string
  cost: number
  tokens: number
  requests: number
}

export interface TopConsumer {
  memberId: string
  memberName: string
  department: string
  cost: number
  tokens: number
  requests: number
}

export interface ModelUsage {
  modelId: string
  modelName: string
  provider: ProviderType
  requests: number
  tokens: number
  cost: number
  percentage: number
}

export interface TeamUsage {
  departmentId: string
  departmentName: string
  quota: number
  consumed: number
  memberCount: number
  topModel: string
}

// ========== 审计日志 ==========

export type AuditAction =
  | 'key_create'
  | 'key_disable'
  | 'key_rotate'
  | 'budget_change'
  | 'budget_approve'
  | 'permission_change'
  | 'role_assign'
  | 'model_whitelist_change'
  | 'member_add'
  | 'member_remove'
  | 'org_structure_change'

export interface OperationLog {
  id: string
  action: AuditAction
  operator: string
  operatorId: string
  target: string
  detail: string
  ip: string
  createdAt: string
}

export interface CallLog {
  id: string
  caller: string
  callerId: string
  callerType: 'member' | 'platform_key'
  model: string
  provider: ProviderType
  inputTokens: number
  outputTokens: number
  latencyMs: number
  status: 'success' | 'error' | 'filtered'
  cost: number
  createdAt: string
  inputPreview: string
  outputPreview: string
}

export interface AuditSettings {
  contentRetentionEnabled: boolean
}

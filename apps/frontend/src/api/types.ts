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
  clientId: string
  clientSecret: string
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
}

// 同步记录
export interface SyncLog {
  id: string
  time: string
  type: 'scheduled' | 'manual'
  result: 'success' | 'partial_failure' | 'failure'
  detail: string
}

// 字段映射
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
  username: string
  employeeId: string
  jobTitle: string
  hireDate: string
  departmentId: string
  departmentName: string
  status: MemberStatus
  roles: string[]
  source: 'imported' | 'manual' | 'invited'
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

// 通用分页
export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

// ========== 预算管理 ==========

export type OverrunPolicy = 'hard_reject' | 'approval' | 'downgrade'

export interface BudgetNode {
  id: string
  name: string
  parentId: string | null
  budget: number          // 分配预算（元）
  consumed: number        // 已消耗（元）
  reserved: number        // 预留池（元）
  memberQuota: number     // 平均额度/人（元）
  children?: BudgetNode[]
  overrunPolicy: OverrunPolicy
  period: string          // e.g. '2026-06'
}

export interface BudgetProject {
  id: string
  name: string
  departmentId: string
  departmentName: string
  budget: number
  consumed: number
  memberIds: string[]
  overrunPolicy: OverrunPolicy
  period: string
}

export interface BudgetApproval {
  id: string
  applicantId: string
  applicantName: string
  departmentId: string
  departmentName: string
  amount: number
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  rejectReason?: string
  createdAt: string
  resolvedAt?: string
}

export interface AlertRule {
  id: string
  targetType: 'team' | 'project'
  targetId: string
  targetName: string
  thresholds: number[]    // e.g. [80, 90, 100]
  notifyRoleIds: string[]
  enabled: boolean
}

// ========== API-KEY 管理 ==========

export type ProviderType = 'openai' | 'anthropic' | 'deepseek' | 'qwen' | 'zhipu' | 'baichuan' | 'minimax' | 'moonshot' | 'stepfun' | 'baidu' | 'hunyuan' | 'doubao' | 'sensetime' | 'custom'
export type KeyStatus = 'active' | 'disabled' | 'expired' | 'error'

export interface ProviderKey {
  id: string
  provider: ProviderType
  name: string
  keyPrefix: string       // 仅展示前缀
  status: KeyStatus
  balance: number | null
  lastUsed: string | null
  createdAt: string
  rotateEnabled: boolean
}

export type QuotaMode = 'fixed' | 'periodic'

export interface PlatformKey {
  id: string
  name: string
  keyPrefix: string
  type: 'member' | 'project'
  memberId: string | null
  memberName: string | null
  projectId: string | null
  projectName: string | null
  departmentId: string
  departmentName: string
  status: KeyStatus
  quotaMode: QuotaMode
  quota: number
  used: number
  modelWhitelist: string[]
  createdAt: string
  expiresAt: string | null
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'

export interface KeyApproval {
  id: string
  applicant: string
  applicantId: string
  department: string
  reason: string
  requestedQuota: number
  requestedModels: string[]
  status: ApprovalStatus
  approver: string | null
  createdAt: string
  resolvedAt: string | null
}

// ========== 模型路由 ==========

export type ModelType = 'builtin' | 'custom'
export type AuthMethod = 'api_key' | 'ak_sk'
export type Visibility = 'all' | 'partial'

export interface ModelInfo {
  id: string
  provider: ProviderType
  name: string            // 模型 ID（用于请求）
  displayName: string     // 展示名称
  type: ModelType
  description: string
  inputPrice: number      // 每百万 token 价格
  outputPrice: number
  maxContext: number
  maxOutput: number
  enabled: boolean
  capabilities: string[]  // e.g. ['chat', 'vision', 'function_calling']
  // 自定义模型专用
  endpoint?: string       // 模型部署地址
  authMethod?: AuthMethod
  apiKey?: string
  proxyUrl?: string
  visibility: Visibility
  visibleDepartmentIds?: string[]
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
  monthOverMonth: number  // 环比百分比
  totalTokens: number
  totalRequests: number
  avgCostPerRequest: number
}

export interface DepartmentCost {
  departmentId: string
  departmentName: string
  cost: number
  percentage: number
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
  | 'key_create' | 'key_disable' | 'key_rotate'
  | 'budget_change' | 'budget_approve'
  | 'permission_change' | 'role_assign'
  | 'model_whitelist_change' | 'model_create' | 'model_update' | 'model_delete'
  | 'alert_create' | 'alert_update' | 'alert_delete'
  | 'member_add' | 'member_remove'
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

// ========== 钱包管理 ==========

export type PaymentMethod = 'alipay' | 'wechat'
export type TopUpStatus = 'success' | 'pending' | 'failed'
export type InvoiceStatus = 'none' | 'applied' | 'issued'

export interface WalletSummary {
  balance: number
  totalConsumed: number
  totalRequests: number
  invitedCount: number
}

export interface TopUpRecord {
  id: string
  orderId: string
  method: PaymentMethod
  amount: number
  paidAmount: number
  invoiceStatus: InvoiceStatus
  status: TopUpStatus
  createdAt: string
}

export interface ReferralInfo {
  pendingReward: number
  totalReward: number
  invitedCount: number
  referralLink: string
  referralCode: string
}

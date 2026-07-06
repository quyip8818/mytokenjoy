import type {
  AlertRule,
  BudgetApproval,
  BudgetNode,
  BudgetProject,
  CallLog,
  CostSummary,
  DailyCost,
  DataSourceStatus,
  Department,
  DepartmentCost,
  FieldMapping,
  KeyApproval,
  Member,
  ModelInfo,
  ModelUsage,
  OperationLog,
  Permission,
  PlatformKey,
  ProviderKey,
  Role,
  RoutingRule,
  SyncConfig,
  SyncLog,
  TeamUsage,
  TopConsumer,
} from '../api/types'

export const mockDataSourceStatus: DataSourceStatus = {
  platform: null,
  connected: false,
  lastImport: null,
  lastImportResult: null,
}

export const mockSyncConfig: SyncConfig = {
  enabled: false,
  startTime: '02:00',
  frequencyHours: 12,
  deleteMemberThreshold: 10,
  deleteDepartmentThreshold: 5,
  notifyPhone: true,
  notifyEmail: true,
}

export const mockSyncLogs: SyncLog[] = [
  { id: '1', time: '2026-06-15 14:00', type: 'scheduled', result: 'success', detail: '新增 2 人' },
  { id: '2', time: '2026-06-15 02:00', type: 'scheduled', result: 'partial_failure', detail: '成功 118 人，失败 2 人' },
  { id: '3', time: '2026-06-14 14:00', type: 'manual', result: 'success', detail: '无变更' },
]

export const mockFieldMappings: FieldMapping[] = [
  { sourceField: 'user_name', sourceLabel: '用户姓名', targetField: 'name', required: true },
  { sourceField: 'mobile', sourceLabel: '手机号码', targetField: 'phone', required: true },
  { sourceField: 'user_email', sourceLabel: '邮箱地址', targetField: 'email', required: false },
  { sourceField: 'dept_name', sourceLabel: '部门名称', targetField: 'departmentName', required: true },
  { sourceField: 'dept_id', sourceLabel: '部门 ID', targetField: 'departmentId', required: true },
  { sourceField: 'user_status', sourceLabel: '用户状态', targetField: 'status', required: false },
]

export const mockDepartments: Department[] = [
  {
    id: 'dept-1',
    name: '总公司',
    parentId: null,
    memberCount: 128,
    children: [
      {
        id: 'dept-2',
        name: '技术部',
        parentId: 'dept-1',
        memberCount: 45,
        children: [
          { id: 'dept-3', name: '后端组', parentId: 'dept-2', memberCount: 20 },
          { id: 'dept-4', name: '前端组', parentId: 'dept-2', memberCount: 15 },
          { id: 'dept-5', name: '测试组', parentId: 'dept-2', memberCount: 10 },
        ],
      },
      { id: 'dept-6', name: '产品部', parentId: 'dept-1', memberCount: 25 },
      { id: 'dept-7', name: '市场部', parentId: 'dept-1', memberCount: 30 },
      { id: 'dept-8', name: '行政部', parentId: 'dept-1', memberCount: 28 },
    ],
  },
]

export const mockMembers: Member[] = [
  { id: 'm-1', name: '张三', phone: '13812341234', email: 'zhangsan@example.com', username: 'zhangsan', employeeId: 'EMP001', jobTitle: '高级后端工程师', hireDate: '2023-03-15', departmentId: 'dept-3', departmentName: '后端组', status: 'active', roles: ['普通成员'], source: 'imported' },
  { id: 'm-2', name: '李四', phone: '13912345678', email: 'lisi@example.com', username: 'lisi', employeeId: 'EMP002', jobTitle: '技术经理', hireDate: '2022-06-01', departmentId: 'dept-3', departmentName: '后端组', status: 'active', roles: ['普通成员', '组织管理员'], source: 'imported' },
  { id: 'm-3', name: '王五', phone: '', email: 'wangwu@example.com', username: 'wangwu', employeeId: 'EMP003', jobTitle: '后端工程师', hireDate: '2024-09-10', departmentId: 'dept-3', departmentName: '后端组', status: 'pending', roles: ['普通成员'], source: 'invited' },
  { id: 'm-4', name: '赵六', phone: '13712349876', email: 'zhaoliu@example.com', username: 'zhaoliu', employeeId: 'EMP004', jobTitle: '前端工程师', hireDate: '2023-07-20', departmentId: 'dept-4', departmentName: '前端组', status: 'active', roles: ['普通成员'], source: 'manual' },
  { id: 'm-5', name: '钱七', phone: '13612340000', email: 'qianqi@example.com', username: 'qianqi', employeeId: 'EMP005', jobTitle: '高级前端工程师', hireDate: '2022-11-05', departmentId: 'dept-4', departmentName: '前端组', status: 'inactive', roles: ['普通成员'], source: 'imported' },
  { id: 'm-6', name: '孙八', phone: '13512348888', email: 'sunba@example.com', username: 'sunba', employeeId: 'EMP006', jobTitle: '测试工程师', hireDate: '2023-05-08', departmentId: 'dept-5', departmentName: '测试组', status: 'active', roles: ['普通成员'], source: 'imported' },
  { id: 'm-7', name: '周九', phone: '13412347777', email: 'zhoujiu@example.com', username: 'zhoujiu', employeeId: 'EMP007', jobTitle: '高级测试工程师', hireDate: '2022-08-12', departmentId: 'dept-5', departmentName: '测试组', status: 'active', roles: ['普通成员'], source: 'imported' },
  { id: 'm-8', name: '吴十', phone: '13312346666', email: 'wushi@example.com', username: 'wushi', employeeId: 'EMP008', jobTitle: '产品经理', hireDate: '2023-01-10', departmentId: 'dept-6', departmentName: '产品部', status: 'active', roles: ['普通成员'], source: 'imported' },
  { id: 'm-9', name: '郑十一', phone: '13212345555', email: 'zheng11@example.com', username: 'zheng11', employeeId: 'EMP009', jobTitle: '产品助理', hireDate: '2025-02-20', departmentId: 'dept-6', departmentName: '产品部', status: 'pending', roles: ['普通成员'], source: 'invited' },
  { id: 'm-10', name: '冯十二', phone: '13112344444', email: 'feng12@example.com', username: 'feng12', employeeId: 'EMP010', jobTitle: '市场总监', hireDate: '2021-04-01', departmentId: 'dept-7', departmentName: '市场部', status: 'active', roles: ['普通成员'], source: 'imported' },
  { id: 'm-11', name: '陈十三', phone: '13012343333', email: 'chen13@example.com', username: 'chen13', employeeId: 'EMP011', jobTitle: '市场专员', hireDate: '2024-03-18', departmentId: 'dept-7', departmentName: '市场部', status: 'active', roles: ['普通成员'], source: 'imported' },
  { id: 'm-12', name: '褚十四', phone: '15912342222', email: 'chu14@example.com', username: 'chu14', employeeId: 'EMP012', jobTitle: '行政主管', hireDate: '2022-02-14', departmentId: 'dept-8', departmentName: '行政部', status: 'active', roles: ['普通成员'], source: 'manual' },
  { id: 'm-13', name: '卫十五', phone: '15812341111', email: 'wei15@example.com', username: 'wei15', employeeId: 'EMP013', jobTitle: '技术总监', hireDate: '2021-01-05', departmentId: 'dept-2', departmentName: '技术部', status: 'active', roles: ['组织管理员'], source: 'imported' },
  { id: 'm-14', name: '蒋十六', phone: '15712340000', email: 'jiang16@example.com', username: 'jiang16', employeeId: 'EMP014', jobTitle: '后端工程师', hireDate: '2025-06-01', departmentId: 'dept-3', departmentName: '后端组', status: 'pending', roles: ['普通成员'], source: 'invited' },
  { id: 'm-15', name: '沈十七', phone: '15612349999', email: 'shen17@example.com', username: 'shen17', employeeId: 'EMP015', jobTitle: '前端工程师', hireDate: '2024-01-15', departmentId: 'dept-4', departmentName: '前端组', status: 'active', roles: ['普通成员'], source: 'imported' },
]

export const mockRoles: Role[] = [
  { id: 'role-1', name: '超级管理员', type: 'preset', permissions: ['*'], memberCount: 1 },
  { id: 'role-2', name: '组织管理员', type: 'preset', permissions: ['org:*'], memberCount: 2 },
  { id: 'role-3', name: '普通成员', type: 'preset', permissions: ['self:*'], memberCount: 128 },
  { id: 'role-4', name: '只读审计员', type: 'preset', permissions: ['audit:read'], memberCount: 3 },
  { id: 'role-5', name: 'API 调用者', type: 'preset', permissions: ['api:call'], memberCount: 50 },
  { id: 'role-6', name: '预算审批员', type: 'custom', permissions: ['budget:approve', 'budget:read'], memberCount: 2 },
]

export const mockPermissions: Permission[] = [
  { id: 'p-1', name: '组织架构管理', group: '组织' },
  { id: 'p-2', name: '成员管理', group: '组织' },
  { id: 'p-3', name: '角色管理', group: '组织' },
  { id: 'p-4', name: '数据源配置', group: '组织' },
  { id: 'p-5', name: '预算分配', group: '资源管控' },
  { id: 'p-6', name: '预算审批', group: '资源管控' },
  { id: 'p-7', name: '模型白名单', group: '资源管控' },
  { id: 'p-8', name: '查看成本看板', group: '运营' },
  { id: 'p-9', name: '用量分析', group: '运营' },
  { id: 'p-10', name: '审计日志查看', group: '运营' },
  { id: 'p-11', name: 'API 调用', group: 'API' },
]

// ========== 预算管理 Mock ==========

export const mockBudgetTree: BudgetNode[] = [
  {
    id: 'dept-1', name: '总公司', parentId: null, budget: 100000, consumed: 67500, reserved: 5000, memberQuota: 0, overrunPolicy: 'approval', period: '2026-06',
    children: [
      {
        id: 'dept-2', name: '技术部', parentId: 'dept-1', budget: 50000, consumed: 38200, reserved: 3000, memberQuota: 1000, overrunPolicy: 'approval', period: '2026-06',
        children: [
          { id: 'dept-3', name: '后端组', parentId: 'dept-2', budget: 25000, consumed: 21000, reserved: 1500, memberQuota: 800, overrunPolicy: 'hard_reject', period: '2026-06' },
          { id: 'dept-4', name: '前端组', parentId: 'dept-2', budget: 15000, consumed: 11200, reserved: 1000, memberQuota: 1000, overrunPolicy: 'downgrade', period: '2026-06' },
          { id: 'dept-5', name: '测试组', parentId: 'dept-2', budget: 10000, consumed: 6000, reserved: 500, memberQuota: 600, overrunPolicy: 'approval', period: '2026-06' },
        ],
      },
      { id: 'dept-6', name: '产品部', parentId: 'dept-1', budget: 20000, consumed: 14300, reserved: 2000, memberQuota: 1200, overrunPolicy: 'downgrade', period: '2026-06' },
      { id: 'dept-7', name: '市场部', parentId: 'dept-1', budget: 15000, consumed: 8500, reserved: 1000, memberQuota: 800, overrunPolicy: 'hard_reject', period: '2026-06' },
      { id: 'dept-8', name: '行政部', parentId: 'dept-1', budget: 15000, consumed: 6500, reserved: 1000, memberQuota: 500, overrunPolicy: 'hard_reject', period: '2026-06' },
    ],
  },
]

export const mockBudgetProjects: BudgetProject[] = [
  { id: 'proj-1', name: 'AI 搜索项目', departmentId: 'dept-3', departmentName: '后端组', budget: 8000, consumed: 5200, memberIds: ['m-1', 'm-4'], overrunPolicy: 'approval', period: '2026-06' },
  { id: 'proj-2', name: '客服 Bot', departmentId: 'dept-6', departmentName: '产品部', budget: 5000, consumed: 3100, memberIds: ['m-2', 'm-8'], overrunPolicy: 'hard_reject', period: '2026-06' },
  { id: 'proj-3', name: '智能测试平台', departmentId: 'dept-5', departmentName: '测试组', budget: 3000, consumed: 1800, memberIds: ['m-6', 'm-7'], overrunPolicy: 'downgrade', period: '2026-06' },
]

export const mockBudgetApprovals: BudgetApproval[] = [
  { id: 'appr-1', applicantId: 'm-1', applicantName: '张三', departmentId: 'dept-3', departmentName: '后端组', amount: 500, reason: '本月额度用尽，需完成搜索优化任务', status: 'pending', createdAt: '2026-06-28 14:30' },
  { id: 'appr-1b', applicantId: 'm-1', applicantName: '张三', departmentId: 'dept-3', departmentName: '后端组', amount: 300, reason: 'RAG 管道调试需额外调用', status: 'approved', createdAt: '2026-06-20 09:00', resolvedAt: '2026-06-20 11:30' },
  { id: 'appr-1c', applicantId: 'm-1', applicantName: '张三', departmentId: 'dept-3', departmentName: '后端组', amount: 200, reason: '紧急修复线上搜索问题', status: 'approved', createdAt: '2026-06-15 16:00', resolvedAt: '2026-06-15 16:45' },
  { id: 'appr-2', applicantId: 'm-4', applicantName: '赵六', departmentId: 'dept-3', departmentName: '后端组', amount: 300, reason: '调试 RAG 管道需额外调用', status: 'pending', createdAt: '2026-06-29 09:15' },
  { id: 'appr-3', applicantId: 'm-8', applicantName: '吴十', departmentId: 'dept-6', departmentName: '产品部', amount: 200, reason: '产品文档生成', status: 'approved', createdAt: '2026-06-25 16:00', resolvedAt: '2026-06-25 17:30' },
]

export const mockAlertRules: AlertRule[] = [
  { id: 'alert-1', targetType: 'team', targetId: 'dept-1', targetName: '总公司', thresholds: [80, 90, 100], notifyRoleIds: ['role-1'], enabled: true },
  { id: 'alert-2', targetType: 'team', targetId: 'dept-2', targetName: '技术部', thresholds: [80, 90, 100], notifyRoleIds: ['role-2'], enabled: true },
  { id: 'alert-3', targetType: 'team', targetId: 'dept-3', targetName: '后端组', thresholds: [90, 100], notifyRoleIds: ['role-2'], enabled: true },
  { id: 'alert-4', targetType: 'project', targetId: 'proj-1', targetName: 'AI 搜索项目', thresholds: [80, 100], notifyRoleIds: ['role-6'], enabled: true },
  { id: 'alert-5', targetType: 'team', targetId: 'dept-6', targetName: '产品部', thresholds: [80, 100], notifyRoleIds: ['role-6'], enabled: false },
]

// ========== API-KEY Mock ==========

export const mockProviderKeys: ProviderKey[] = [
  { id: 'pk-1', provider: 'openai', name: 'OpenAI 主力', keyPrefix: 'sk-proj-abc...', status: 'active', balance: 4250.00, lastUsed: '2026-06-19 10:32', createdAt: '2026-01-15', rotateEnabled: true },
  { id: 'pk-2', provider: 'anthropic', name: 'Anthropic 生产', keyPrefix: 'sk-ant-xyz...', status: 'active', balance: 2100.00, lastUsed: '2026-06-19 09:45', createdAt: '2026-02-01', rotateEnabled: true },
  { id: 'pk-3', provider: 'deepseek', name: 'DeepSeek V3', keyPrefix: 'sk-ds-mno...', status: 'active', balance: 800.00, lastUsed: '2026-06-18 16:20', createdAt: '2026-03-10', rotateEnabled: false },
  { id: 'pk-4', provider: 'qwen', name: '通义千问', keyPrefix: 'sk-qw-pqr...', status: 'disabled', balance: null, lastUsed: '2026-05-20 12:00', createdAt: '2026-04-01', rotateEnabled: false },
  { id: 'pk-5', provider: 'openai', name: 'OpenAI 备用', keyPrefix: 'sk-proj-def...', status: 'active', balance: 1500.00, lastUsed: '2026-06-17 08:00', createdAt: '2026-05-01', rotateEnabled: true },
]

export const mockPlatformKeys: PlatformKey[] = [
  { id: 'plk-1', name: '开发调试', keyPrefix: 'tj-dev-001...', type: 'member', memberId: 'm-1', memberName: '张三', projectId: null, projectName: null, departmentId: 'dept-3', departmentName: '后端组', status: 'active', quotaMode: 'periodic', quota: 5000, used: 3200, modelWhitelist: ['gpt-4o', 'claude-sonnet-4-6'], createdAt: '2026-03-01', expiresAt: '2026-12-31' },
  { id: 'plk-1b', name: 'RAG 管道测试', keyPrefix: 'tj-rag-001b...', type: 'member', memberId: 'm-1', memberName: '张三', projectId: null, projectName: null, departmentId: 'dept-3', departmentName: '后端组', status: 'active', quotaMode: 'periodic', quota: 3000, used: 1800, modelWhitelist: ['deepseek-v3', 'claude-sonnet-4-6'], createdAt: '2026-05-15', expiresAt: null },
  { id: 'plk-1c', name: '搜索优化实验', keyPrefix: 'tj-srch-001c...', type: 'member', memberId: 'm-1', memberName: '张三', projectId: null, projectName: null, departmentId: 'dept-3', departmentName: '后端组', status: 'disabled', quotaMode: 'fixed', quota: 1000, used: 1000, modelWhitelist: ['gpt-4o-mini'], createdAt: '2026-04-10', expiresAt: '2026-06-01' },
  { id: 'plk-2', name: '后端服务调用', keyPrefix: 'tj-svc-002...', type: 'member', memberId: 'm-2', memberName: '李四', projectId: null, projectName: null, departmentId: 'dept-3', departmentName: '后端组', status: 'active', quotaMode: 'periodic', quota: 10000, used: 7800, modelWhitelist: ['gpt-4o', 'deepseek-v3'], createdAt: '2026-03-15', expiresAt: null },
  { id: 'plk-3', name: '前端实验', keyPrefix: 'tj-exp-003...', type: 'member', memberId: 'm-4', memberName: '赵六', projectId: null, projectName: null, departmentId: 'dept-4', departmentName: '前端组', status: 'disabled', quotaMode: 'fixed', quota: 2000, used: 2000, modelWhitelist: ['gpt-4o-mini'], createdAt: '2026-05-01', expiresAt: '2026-06-30' },
  { id: 'plk-4', name: '产品文档生成', keyPrefix: 'tj-doc-004...', type: 'member', memberId: 'm-8', memberName: '吴十', projectId: null, projectName: null, departmentId: 'dept-6', departmentName: '产品部', status: 'active', quotaMode: 'periodic', quota: 3000, used: 1200, modelWhitelist: ['gpt-4o-mini', 'deepseek-v3'], createdAt: '2026-04-10', expiresAt: null },
  { id: 'plk-5', name: '自动化脚本', keyPrefix: 'tj-auto-005...', type: 'member', memberId: 'm-13', memberName: '卫十五', projectId: null, projectName: null, departmentId: 'dept-2', departmentName: '技术部', status: 'active', quotaMode: 'periodic', quota: 15000, used: 9200, modelWhitelist: ['gpt-4o', 'claude-sonnet-4-6', 'deepseek-v3'], createdAt: '2026-02-01', expiresAt: null },
  { id: 'plk-6', name: 'prod-搜索服务', keyPrefix: 'tj-srch-006...', type: 'project', memberId: null, memberName: null, projectId: 'proj-1', projectName: 'AI 搜索项目', departmentId: 'dept-3', departmentName: '后端组', status: 'active', quotaMode: 'periodic', quota: 20000, used: 15600, modelWhitelist: ['gpt-4o', 'deepseek-v3'], createdAt: '2026-04-01', expiresAt: null },
  { id: 'plk-7', name: 'dev-搜索测试', keyPrefix: 'tj-srch-007...', type: 'project', memberId: null, memberName: null, projectId: 'proj-1', projectName: 'AI 搜索项目', departmentId: 'dept-3', departmentName: '后端组', status: 'active', quotaMode: 'fixed', quota: 5000, used: 2100, modelWhitelist: ['gpt-4o', 'deepseek-v3'], createdAt: '2026-05-10', expiresAt: '2026-09-30' },
  { id: 'plk-8', name: 'prod-客服对话', keyPrefix: 'tj-bot-008...', type: 'project', memberId: null, memberName: null, projectId: 'proj-2', projectName: '客服 Bot', departmentId: 'dept-6', departmentName: '产品部', status: 'active', quotaMode: 'periodic', quota: 8000, used: 4500, modelWhitelist: ['gpt-4o-mini', 'deepseek-v3'], createdAt: '2026-03-20', expiresAt: null },
  { id: 'plk-9', name: 'prod-测试执行', keyPrefix: 'tj-test-009...', type: 'project', memberId: null, memberName: null, projectId: 'proj-3', projectName: '智能测试平台', departmentId: 'dept-5', departmentName: '测试组', status: 'active', quotaMode: 'periodic', quota: 6000, used: 3800, modelWhitelist: ['claude-sonnet-4-6', 'deepseek-v3'], createdAt: '2026-04-15', expiresAt: null },
]
export const mockApprovals: KeyApproval[] = [
  { id: 'apv-1', applicant: '钱七', applicantId: 'm-5', department: '前端组', reason: '需要接入 GPT-4o 进行代码辅助开发', requestedQuota: 5000, requestedModels: ['gpt-4o', 'claude-sonnet-4-6'], status: 'pending', approver: null, createdAt: '2026-06-18 14:30', resolvedAt: null },
  { id: 'apv-2', applicant: '王五', applicantId: 'm-3', department: '后端组', reason: '新项目需要多模型测试', requestedQuota: 8000, requestedModels: ['gpt-4o', 'deepseek-v3', 'claude-sonnet-4-6'], status: 'pending', approver: null, createdAt: '2026-06-17 09:15', resolvedAt: null },
  { id: 'apv-3', applicant: '张三', applicantId: 'm-1', department: '后端组', reason: '额度即将用完，申请追加', requestedQuota: 3000, requestedModels: ['gpt-4o'], status: 'approved', approver: '李四', createdAt: '2026-06-15 11:00', resolvedAt: '2026-06-15 14:20' },
  { id: 'apv-4', applicant: '赵六', applicantId: 'm-4', department: '前端组', reason: '探索 Agent 功能', requestedQuota: 10000, requestedModels: ['gpt-4o', 'claude-opus-4-8'], status: 'rejected', approver: '李四', createdAt: '2026-06-10 16:00', resolvedAt: '2026-06-11 09:30' },
]

// ========== 模型路由 Mock ==========

export const mockModels: ModelInfo[] = [
  { id: 'model-1', provider: 'openai', name: 'gpt-4o', displayName: 'GPT-4o', type: 'builtin', description: '旗舰多模态模型，支持视觉和函数调用', inputPrice: 2.5, outputPrice: 10, maxContext: 128000, maxOutput: 16384, enabled: true, capabilities: ['chat', 'vision', 'function_calling'], visibility: 'all' },
  { id: 'model-2', provider: 'openai', name: 'gpt-4o-mini', displayName: 'GPT-4o Mini', type: 'builtin', description: '轻量高性价比模型，适合日常任务', inputPrice: 0.15, outputPrice: 0.6, maxContext: 128000, maxOutput: 16384, enabled: true, capabilities: ['chat', 'vision', 'function_calling'], visibility: 'all' },
  { id: 'model-3', provider: 'anthropic', name: 'claude-opus-4-8', displayName: 'Claude Opus 4.8', type: 'builtin', description: '最强推理模型，支持 1M 上下文', inputPrice: 15, outputPrice: 75, maxContext: 1000000, maxOutput: 32000, enabled: true, capabilities: ['chat', 'vision', 'function_calling'], visibility: 'all' },
  { id: 'model-4', provider: 'anthropic', name: 'claude-sonnet-4-6', displayName: 'Claude Sonnet 4.6', type: 'builtin', description: '平衡性能与成本的中端模型', inputPrice: 3, outputPrice: 15, maxContext: 200000, maxOutput: 8192, enabled: true, capabilities: ['chat', 'vision', 'function_calling'], visibility: 'all' },
  { id: 'model-5', provider: 'deepseek', name: 'deepseek-v3', displayName: 'DeepSeek V3', type: 'builtin', description: '高性价比开源模型', inputPrice: 0.27, outputPrice: 1.1, maxContext: 64000, maxOutput: 8192, enabled: true, capabilities: ['chat', 'function_calling'], visibility: 'all' },
  { id: 'model-6', provider: 'deepseek', name: 'deepseek-r1', displayName: 'DeepSeek R1', type: 'builtin', description: '推理增强模型', inputPrice: 0.55, outputPrice: 2.19, maxContext: 64000, maxOutput: 8192, enabled: false, capabilities: ['chat', 'function_calling'], visibility: 'all' },
  { id: 'model-7', provider: 'qwen', name: 'qwen-max', displayName: 'Qwen Max', type: 'builtin', description: '通义千问旗舰模型', inputPrice: 2.0, outputPrice: 6.0, maxContext: 32000, maxOutput: 8192, enabled: false, capabilities: ['chat', 'function_calling'], visibility: 'all' },
  { id: 'model-8', provider: 'qwen', name: 'qwen-plus', displayName: 'Qwen Plus', type: 'builtin', description: '通义千问增强版，支持视觉', inputPrice: 0.8, outputPrice: 2.0, maxContext: 131072, maxOutput: 8192, enabled: true, capabilities: ['chat', 'vision'], visibility: 'all' },
  { id: 'model-9', provider: 'custom', name: 'hunyuan-pro', displayName: '混元 Pro', type: 'custom', description: '腾讯混元大模型，自部署接入', inputPrice: 1.5, outputPrice: 5.0, maxContext: 128000, maxOutput: 4096, enabled: true, capabilities: ['chat'], endpoint: 'https://api.hunyuan.tencent.com/v1/chat', authMethod: 'api_key', visibility: 'partial', visibleDepartmentIds: ['dept-2'] },
]
export const mockRoutingRules: RoutingRule[] = [
  { id: 'rr-1', nodeId: 'dept-1', nodeName: '总公司', allowedModels: ['gpt-4o', 'gpt-4o-mini', 'claude-sonnet-4-6', 'deepseek-v3', 'qwen-plus'], defaultModel: 'gpt-4o-mini', fallbackModel: 'deepseek-v3', inherited: false },
  { id: 'rr-2', nodeId: 'dept-2', nodeName: '技术部', allowedModels: ['gpt-4o', 'gpt-4o-mini', 'claude-sonnet-4-6', 'claude-opus-4-8', 'deepseek-v3'], defaultModel: 'gpt-4o', fallbackModel: 'deepseek-v3', inherited: false },
  { id: 'rr-3', nodeId: 'dept-3', nodeName: '后端组', allowedModels: ['gpt-4o', 'claude-sonnet-4-6', 'deepseek-v3'], defaultModel: null, fallbackModel: null, inherited: true },
  { id: 'rr-4', nodeId: 'dept-6', nodeName: '产品部', allowedModels: ['gpt-4o-mini', 'deepseek-v3', 'qwen-plus'], defaultModel: 'gpt-4o-mini', fallbackModel: 'qwen-plus', inherited: false },
]

// ========== 数据看板 Mock ==========

export const mockCostSummary: CostSummary = {
  totalCost: 67500,
  monthOverMonth: 12.5,
  totalTokens: 45000000,
  totalRequests: 28500,
  avgCostPerRequest: 2.37,
}

export const mockDepartmentCosts: DepartmentCost[] = [
  { departmentId: 'dept-2', departmentName: '技术部', cost: 38200, percentage: 56.6 },
  { departmentId: 'dept-6', departmentName: '产品部', cost: 14300, percentage: 21.2 },
  { departmentId: 'dept-7', departmentName: '市场部', cost: 8500, percentage: 12.6 },
  { departmentId: 'dept-8', departmentName: '行政部', cost: 6500, percentage: 9.6 },
]

export const mockDailyCosts: DailyCost[] = Array.from({ length: 30 }, (_, i) => {
  const date = new Date(2026, 5, i + 1)
  const base = 2000 + Math.random() * 1500
  return {
    date: date.toISOString().split('T')[0],
    cost: Math.round(base * 100) / 100,
    tokens: Math.round(base * 700),
    requests: Math.round(base / 2.5),
  }
})

export const mockTopConsumers: TopConsumer[] = [
  { memberId: 'm-2', memberName: '李四', department: '后端组', cost: 12500, tokens: 8500000, requests: 5200 },
  { memberId: 'm-1', memberName: '张三', department: '后端组', cost: 8700, tokens: 5800000, requests: 3800 },
  { memberId: 'm-4', memberName: '赵六', department: '前端组', cost: 6200, tokens: 4100000, requests: 2900 },
  { memberId: 'm-5', memberName: '钱七', department: '前端组', cost: 5000, tokens: 3300000, requests: 2100 },
  { memberId: 'm-3', memberName: '王五', department: '后端组', cost: 3800, tokens: 2500000, requests: 1500 },
]

export const mockModelUsage: ModelUsage[] = [
  { modelId: 'model-1', modelName: 'GPT-4o', provider: 'openai', requests: 12000, tokens: 18000000, cost: 32000, percentage: 47.4 },
  { modelId: 'model-5', modelName: 'DeepSeek V3', provider: 'deepseek', requests: 8500, tokens: 15000000, cost: 12500, percentage: 18.5 },
  { modelId: 'model-4', modelName: 'Claude Sonnet 4.6', provider: 'anthropic', requests: 4500, tokens: 7000000, cost: 14000, percentage: 20.7 },
  { modelId: 'model-2', modelName: 'GPT-4o Mini', provider: 'openai', requests: 3000, tokens: 4000000, cost: 5500, percentage: 8.1 },
  { modelId: 'model-8', modelName: 'Qwen Plus', provider: 'qwen', requests: 500, tokens: 1000000, cost: 3500, percentage: 5.2 },
]

export const mockTeamUsage: TeamUsage[] = [
  { departmentId: 'dept-3', departmentName: '后端组', quota: 25000, consumed: 21000, memberCount: 20, topModel: 'GPT-4o' },
  { departmentId: 'dept-4', departmentName: '前端组', quota: 15000, consumed: 11200, memberCount: 15, topModel: 'Claude Sonnet 4.6' },
  { departmentId: 'dept-5', departmentName: '测试组', quota: 10000, consumed: 6000, memberCount: 10, topModel: 'DeepSeek V3' },
  { departmentId: 'dept-6', departmentName: '产品部', quota: 20000, consumed: 14300, memberCount: 25, topModel: 'GPT-4o Mini' },
  { departmentId: 'dept-7', departmentName: '市场部', quota: 15000, consumed: 8500, memberCount: 30, topModel: 'Qwen Plus' },
]

// ========== 审计日志 Mock ==========

export const mockOperationLogs: OperationLog[] = [
  { id: 'op-1', action: 'key_create', operator: '李四', operatorId: 'm-2', target: 'Platform Key: 客服Bot-生产', detail: '创建平台凭证，额度 20000 元', ip: '192.168.1.100', createdAt: '2026-06-19 10:30' },
  { id: 'op-2', action: 'budget_change', operator: '李四', operatorId: 'm-2', target: '后端组', detail: '预算从 20000 调整为 25000', ip: '192.168.1.100', createdAt: '2026-06-19 09:15' },
  { id: 'op-3', action: 'budget_approve', operator: '李四', operatorId: 'm-2', target: '张三', detail: '审批通过额度追加 3000 元', ip: '192.168.1.100', createdAt: '2026-06-18 14:20' },
  { id: 'op-4', action: 'key_disable', operator: '李四', operatorId: 'm-2', target: 'Platform Key: 赵六-实验', detail: '额度已耗尽，自动禁用', ip: 'system', createdAt: '2026-06-18 11:00' },
  { id: 'op-5', action: 'model_create', operator: '李四', operatorId: 'm-2', target: '混元 Pro', detail: '新增自定义模型，部署地址 https://api.hunyuan.tencent.com/v1/chat', ip: '192.168.1.100', createdAt: '2026-06-17 16:45' },
  { id: 'op-6', action: 'permission_change', operator: '张三', operatorId: 'm-1', target: '钱七', detail: '授予 API 调用者 角色', ip: '192.168.1.105', createdAt: '2026-06-17 10:00' },
  { id: 'op-7', action: 'role_assign', operator: '李四', operatorId: 'm-2', target: '预算审批员', detail: '添加成员：王五', ip: '192.168.1.100', createdAt: '2026-06-16 14:30' },
  { id: 'op-8', action: 'key_rotate', operator: '李四', operatorId: 'm-2', target: 'OpenAI 主力', detail: 'Provider Key 轮转成功', ip: '192.168.1.100', createdAt: '2026-06-15 08:00' },
  { id: 'op-9', action: 'org_structure_change', operator: '李四', operatorId: 'm-2', target: '技术部', detail: '新增子部门：AI 实验室', ip: '192.168.1.100', createdAt: '2026-06-14 11:20' },
  { id: 'op-10', action: 'member_add', operator: '张三', operatorId: 'm-1', target: '后端组', detail: '邀请成员：王五', ip: '192.168.1.105', createdAt: '2026-06-13 15:00' },
  { id: 'op-11', action: 'alert_create', operator: '李四', operatorId: 'm-2', target: '技术部', detail: '创建预警规则：阈值 80%, 90%, 100%', ip: '192.168.1.100', createdAt: '2026-06-13 10:30' },
  { id: 'op-12', action: 'model_update', operator: '李四', operatorId: 'm-2', target: '混元 Pro', detail: '修改模型可见范围为部分成员', ip: '192.168.1.100', createdAt: '2026-06-12 16:00' },
  { id: 'op-13', action: 'model_whitelist_change', operator: '李四', operatorId: 'm-2', target: '技术部', detail: '添加 claude-opus-4-8 到白名单', ip: '192.168.1.100', createdAt: '2026-06-12 11:20' },
  { id: 'op-14', action: 'alert_delete', operator: '张三', operatorId: 'm-1', target: '产品部', detail: '删除预警规则', ip: '192.168.1.105', createdAt: '2026-06-11 09:00' },
  { id: 'op-15', action: 'model_delete', operator: '李四', operatorId: 'm-2', target: '测试模型', detail: '删除自定义模型', ip: '192.168.1.100', createdAt: '2026-06-10 17:30' },
  { id: 'op-16', action: 'budget_change', operator: '张三', operatorId: 'm-1', target: '前端组', detail: '预算从 12000 调整为 15000', ip: '192.168.1.105', createdAt: '2026-06-10 14:00' },
  { id: 'op-17', action: 'member_remove', operator: '李四', operatorId: 'm-2', target: '市场部', detail: '移除成员：临时实习生', ip: '192.168.1.100', createdAt: '2026-06-09 11:00' },
  { id: 'op-18', action: 'key_create', operator: '张三', operatorId: 'm-1', target: 'Provider Key: DeepSeek 备用', detail: '新增供应商 Key', ip: '192.168.1.105', createdAt: '2026-06-08 09:30' },
  { id: 'op-19', action: 'alert_update', operator: '李四', operatorId: 'm-2', target: '总公司', detail: '修改通知角色为超级管理员', ip: '192.168.1.100', createdAt: '2026-06-07 15:45' },
  { id: 'op-20', action: 'budget_approve', operator: '李四', operatorId: 'm-2', target: '赵六', detail: '审批通过额度追加 500 元', ip: '192.168.1.100', createdAt: '2026-06-06 10:20' },
  { id: 'op-21', action: 'org_structure_change', operator: '李四', operatorId: 'm-2', target: '产品部', detail: '部门重命名：产品设计部 → 产品部', ip: '192.168.1.100', createdAt: '2026-06-05 14:00' },
  { id: 'op-22', action: 'key_disable', operator: '李四', operatorId: 'm-2', target: 'Provider Key: 通义千问', detail: '手动禁用', ip: '192.168.1.100', createdAt: '2026-06-04 09:00' },
]

export const mockCallLogs: CallLog[] = [
  { id: 'call-1', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'gpt-4o', provider: 'openai', inputTokens: 1250, outputTokens: 580, latencyMs: 2340, status: 'success', cost: 8.93, createdAt: '2026-06-19 10:32:15', inputPreview: '请帮我优化这段 Go 代码的性能...', outputPreview: '以下是优化建议：1. 使用 sync.Pool...' },
  { id: 'call-2', caller: '智能客服', callerId: 'plk-3', callerType: 'platform_key', model: 'gpt-4o-mini', provider: 'openai', inputTokens: 800, outputTokens: 320, latencyMs: 1120, status: 'success', cost: 0.31, createdAt: '2026-06-19 10:31:42', inputPreview: '用户问题：如何重置密码？', outputPreview: '您可以通过以下步骤重置密码...' },
  { id: 'call-3', caller: '李四', callerId: 'm-2', callerType: 'member', model: 'claude-sonnet-4-6', provider: 'anthropic', inputTokens: 3500, outputTokens: 1200, latencyMs: 4560, status: 'success', cost: 28.5, createdAt: '2026-06-19 10:28:03', inputPreview: '分析这个系统架构的瓶颈...', outputPreview: '根据架构图分析，主要瓶颈在...' },
  { id: 'call-4', caller: '赵六', callerId: 'm-4', callerType: 'member', model: 'deepseek-v3', provider: 'deepseek', inputTokens: 2000, outputTokens: 0, latencyMs: 350, status: 'error', cost: 0.54, createdAt: '2026-06-19 10:25:11', inputPreview: '请生成一个 React 组件...', outputPreview: 'Error: rate_limit_exceeded' },
  { id: 'call-5', caller: '代码审查助手', callerId: 'plk-5', callerType: 'platform_key', model: 'claude-sonnet-4-6', provider: 'anthropic', inputTokens: 5000, outputTokens: 2000, latencyMs: 6200, status: 'success', cost: 45.0, createdAt: '2026-06-19 10:20:00', inputPreview: 'Review this pull request diff...', outputPreview: '## Code Review Summary\n...' },
  { id: 'call-6', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'gpt-4o', provider: 'openai', inputTokens: 900, outputTokens: 450, latencyMs: 1890, status: 'filtered', cost: 0, createdAt: '2026-06-19 10:15:30', inputPreview: '[内容已脱敏]', outputPreview: '[触发合规过滤规则]' },
  { id: 'call-6b', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'claude-sonnet-4-6', provider: 'anthropic', inputTokens: 2800, outputTokens: 1100, latencyMs: 3200, status: 'success', cost: 24.9, createdAt: '2026-06-19 09:45:00', inputPreview: '分析这个搜索索引的设计...', outputPreview: '索引结构可以优化为...' },
  { id: 'call-6c', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'deepseek-v3', provider: 'deepseek', inputTokens: 1500, outputTokens: 600, latencyMs: 1100, status: 'success', cost: 1.07, createdAt: '2026-06-19 09:30:22', inputPreview: '请帮我写一个 Redis 缓存策略...', outputPreview: '推荐使用 Write-Through 策略...' },
  { id: 'call-6d', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'gpt-4o', provider: 'openai', inputTokens: 2100, outputTokens: 800, latencyMs: 2500, status: 'success', cost: 13.25, createdAt: '2026-06-18 17:20:00', inputPreview: 'Review this database migration...', outputPreview: 'Migration looks good, but...' },
  { id: 'call-6e', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'gpt-4o-mini', provider: 'openai', inputTokens: 500, outputTokens: 200, latencyMs: 800, status: 'success', cost: 0.2, createdAt: '2026-06-18 15:10:00', inputPreview: '翻译这段英文文档...', outputPreview: '以下是翻译结果...' },
  { id: 'call-6f', caller: '张三', callerId: 'm-1', callerType: 'member', model: 'deepseek-v3', provider: 'deepseek', inputTokens: 1800, outputTokens: 0, latencyMs: 200, status: 'error', cost: 0.49, createdAt: '2026-06-18 14:00:00', inputPreview: '生成单元测试...', outputPreview: 'Error: context_length_exceeded' },
  { id: 'call-7', caller: '钱七', callerId: 'm-5', callerType: 'member', model: 'gpt-4o-mini', provider: 'openai', inputTokens: 600, outputTokens: 280, latencyMs: 980, status: 'success', cost: 0.26, createdAt: '2026-06-19 10:10:22', inputPreview: '帮我写一个正则表达式...', outputPreview: '这是匹配邮箱的正则：...' },
  { id: 'call-8', caller: '智能客服', callerId: 'plk-3', callerType: 'platform_key', model: 'deepseek-v3', provider: 'deepseek', inputTokens: 1100, outputTokens: 500, latencyMs: 1450, status: 'success', cost: 0.85, createdAt: '2026-06-19 10:05:18', inputPreview: '用户问题：订单状态查询', outputPreview: '您的订单当前状态为...' },
]

// ========== 钱包管理 ==========

export const mockWalletSummary: import('../api/types').WalletSummary = {
  balance: 2.00,
  totalConsumed: 0,
  totalRequests: 0,
  invitedCount: 0,
}

export const mockReferralInfo: import('../api/types').ReferralInfo = {
  pendingReward: 0,
  totalReward: 0,
  invitedCount: 0,
  referralLink: 'https://www.moyu.cn/register?aff=uHPK',
  referralCode: 'uHPK',
}

export const mockTopUpRecords: import('../api/types').TopUpRecord[] = [
  { id: 'tu-1', orderId: 'ORD202606190001', method: 'alipay', amount: 100, paidAmount: 100, invoiceStatus: 'none', status: 'success', createdAt: '2026-06-19 14:30:00' },
  { id: 'tu-2', orderId: 'ORD202606180002', method: 'wechat', amount: 50, paidAmount: 50, invoiceStatus: 'applied', status: 'success', createdAt: '2026-06-18 10:15:00' },
  { id: 'tu-3', orderId: 'ORD202606150003', method: 'alipay', amount: 200, paidAmount: 200, invoiceStatus: 'issued', status: 'success', createdAt: '2026-06-15 09:00:00' },
  { id: 'tu-4', orderId: 'ORD202606120004', method: 'wechat', amount: 20, paidAmount: 20, invoiceStatus: 'none', status: 'pending', createdAt: '2026-06-12 16:45:00' },
  { id: 'tu-5', orderId: 'ORD202606100005', method: 'alipay', amount: 500, paidAmount: 500, invoiceStatus: 'issued', status: 'success', createdAt: '2026-06-10 08:20:00' },
]

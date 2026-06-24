import type {
  DataSourceStatus,
  Department,
  ImportFailure,
  Permission,
  Role,
  SyncConfig,
  SyncLog,
} from '@/api/types'
import {
  buildMockMembers,
  countMembersByRole,
  ROLE_API_CALLER,
  ROLE_AUDITOR,
  ROLE_BUDGET_APPROVER,
  ROLE_MEMBER,
  ROLE_ORG_ADMIN,
  ROLE_SUPER_ADMIN,
} from '../lib/member-factory'

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
  notifyIm: true,
}

export const mockSyncLogs: SyncLog[] = [
  {
    id: 'sync-1',
    time: '2026-06-19 02:00',
    type: 'scheduled',
    result: 'success',
    detail: '新增 3 人，更新 12 人',
  },
  {
    id: 'sync-2',
    time: '2026-06-18 14:00',
    type: 'manual',
    result: 'success',
    detail: '无变更',
  },
  {
    id: 'sync-3',
    time: '2026-06-18 02:00',
    type: 'scheduled',
    result: 'partial_failure',
    detail: '成功 125 人，失败 3 人',
  },
  {
    id: 'sync-4',
    time: '2026-06-17 14:00',
    type: 'scheduled',
    result: 'success',
    detail: '新增 1 人',
  },
  {
    id: 'sync-5',
    time: '2026-06-17 02:00',
    type: 'scheduled',
    result: 'success',
    detail: '部门结构同步完成',
  },
  {
    id: 'sync-6',
    time: '2026-06-16 14:00',
    type: 'manual',
    result: 'failure',
    detail: '数据源连接超时',
  },
  {
    id: 'sync-7',
    time: '2026-06-16 02:00',
    type: 'scheduled',
    result: 'failure',
    detail: '需软删除 15 人，超过保护阈值 10 人，同步已终止并已通知超管',
  },
  {
    id: 'sync-8',
    time: '2026-06-15 14:00',
    type: 'scheduled',
    result: 'success',
    detail: '新增 2 人',
  },
  {
    id: 'sync-9',
    time: '2026-06-15 02:00',
    type: 'scheduled',
    result: 'partial_failure',
    detail: '成功 118 人，失败 2 人',
  },
  {
    id: 'sync-10',
    time: '2026-06-14 14:00',
    type: 'manual',
    result: 'success',
    detail: '无变更',
  },
  {
    id: 'sync-11',
    time: '2026-06-14 02:00',
    type: 'scheduled',
    result: 'success',
    detail: '全量同步完成',
  },
  {
    id: 'sync-12',
    time: '2026-06-13 02:00',
    type: 'scheduled',
    result: 'success',
    detail: '初始化同步 128 人',
  },
]

export const mockImportFailures: ImportFailure[] = [
  { id: 'f-1', name: '李四', employeeId: '10087', reason: '手机号为空' },
  { id: 'f-2', name: '王五', employeeId: '10088', reason: '部门不存在' },
  { id: 'f-3', name: '陈静', employeeId: '10089', reason: '邮箱格式错误' },
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

export const mockMembers = buildMockMembers()

export const mockRoles: Role[] = [
  {
    id: 'role-1',
    name: ROLE_SUPER_ADMIN,
    type: 'preset',
    permissions: ['*'],
    memberCount: countMembersByRole(mockMembers, ROLE_SUPER_ADMIN),
  },
  {
    id: 'role-2',
    name: ROLE_ORG_ADMIN,
    type: 'preset',
    permissions: ['org:*'],
    memberCount: countMembersByRole(mockMembers, ROLE_ORG_ADMIN),
  },
  {
    id: 'role-3',
    name: ROLE_MEMBER,
    type: 'preset',
    permissions: ['self:*'],
    memberCount: countMembersByRole(mockMembers, ROLE_MEMBER),
  },
  {
    id: 'role-4',
    name: ROLE_AUDITOR,
    type: 'preset',
    permissions: ['audit:read'],
    memberCount: countMembersByRole(mockMembers, ROLE_AUDITOR),
  },
  {
    id: 'role-5',
    name: ROLE_API_CALLER,
    type: 'preset',
    permissions: ['api:call'],
    memberCount: countMembersByRole(mockMembers, ROLE_API_CALLER),
  },
  {
    id: 'role-6',
    name: ROLE_BUDGET_APPROVER,
    type: 'custom',
    permissions: ['budget:approve', 'budget:read'],
    memberCount: countMembersByRole(mockMembers, ROLE_BUDGET_APPROVER),
  },
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

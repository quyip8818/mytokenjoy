import type { Member, MemberStatus } from '@/api/types'
import {
  ROLE_API_CALLER,
  ROLE_AUDITOR,
  ROLE_BUDGET_APPROVER,
  ROLE_MEMBER,
  ROLE_ORG_ADMIN,
  ROLE_SUPER_ADMIN,
} from '@/lib/role-constants'
import { buildChineseName, buildEmail, buildPhone } from './names'

export {
  ROLE_API_CALLER,
  ROLE_AUDITOR,
  ROLE_BUDGET_APPROVER,
  ROLE_MEMBER,
  ROLE_ORG_ADMIN,
  ROLE_SUPER_ADMIN,
} from '@/lib/role-constants'

export interface DeptQuota {
  departmentId: string
  departmentName: string
  count: number
}

export const LEAF_DEPT_QUOTAS: DeptQuota[] = [
  { departmentId: 'dept-3', departmentName: '后端组', count: 20 },
  { departmentId: 'dept-4', departmentName: '前端组', count: 15 },
  { departmentId: 'dept-5', departmentName: '测试组', count: 10 },
  { departmentId: 'dept-6', departmentName: '产品部', count: 25 },
  { departmentId: 'dept-7', departmentName: '市场部', count: 30 },
  { departmentId: 'dept-8', departmentName: '行政部', count: 27 },
]

const ANCHOR_MEMBERS: Member[] = [
  {
    id: 'm-admin',
    name: '管理员',
    phone: '13800000001',
    email: 'admin@example.com',
    departmentId: 'dept-1',
    departmentName: '总公司',
    status: 'active',
    roles: [ROLE_SUPER_ADMIN],
    source: 'manual',
  },
  {
    id: 'm-1',
    name: '张三',
    phone: '13812341234',
    email: 'zhangsan@example.com',
    departmentId: 'dept-3',
    departmentName: '后端组',
    status: 'active',
    roles: [ROLE_MEMBER, ROLE_API_CALLER],
    source: 'imported',
  },
  {
    id: 'm-2',
    name: '李四',
    phone: '13912345678',
    email: 'lisi@example.com',
    departmentId: 'dept-3',
    departmentName: '后端组',
    status: 'active',
    roles: [ROLE_MEMBER, ROLE_ORG_ADMIN, ROLE_BUDGET_APPROVER],
    source: 'imported',
  },
  {
    id: 'm-3',
    name: '王五',
    phone: '',
    email: 'wangwu@example.com',
    departmentId: 'dept-3',
    departmentName: '后端组',
    status: 'pending',
    roles: [ROLE_MEMBER],
    source: 'invited',
  },
  {
    id: 'm-4',
    name: '赵六',
    phone: '13712349876',
    email: 'zhaoliu@example.com',
    departmentId: 'dept-4',
    departmentName: '前端组',
    status: 'active',
    roles: [ROLE_MEMBER, ROLE_API_CALLER],
    source: 'manual',
  },
  {
    id: 'm-5',
    name: '钱七',
    phone: '13612340000',
    email: 'qianqi@example.com',
    departmentId: 'dept-4',
    departmentName: '前端组',
    status: 'inactive',
    roles: [ROLE_MEMBER],
    source: 'imported',
  },
  {
    id: 'm-auditor',
    name: '孙审计',
    phone: '13512345678',
    email: 'sunaudit@example.com',
    departmentId: 'dept-8',
    departmentName: '行政部',
    status: 'active',
    roles: [ROLE_MEMBER, ROLE_AUDITOR],
    source: 'manual',
  },
  {
    id: 'm-pure',
    name: '周八',
    phone: '13412345678',
    email: 'zhouba@example.com',
    departmentId: 'dept-3',
    departmentName: '后端组',
    status: 'active',
    roles: [ROLE_MEMBER],
    source: 'manual',
  },
]

function pickStatus(index: number): MemberStatus {
  const mod = index % 100
  if (mod < 7) return 'inactive'
  if (mod < 15) return 'pending'
  return 'active'
}

function pickSource(index: number): Member['source'] {
  const mod = index % 10
  if (mod === 0) return 'invited'
  if (mod <= 2) return 'manual'
  return 'imported'
}

function anchorsInDept(deptId: string): Member[] {
  return ANCHOR_MEMBERS.filter((m) => m.departmentId === deptId)
}

function buildGeneratedMember(
  id: string,
  index: number,
  departmentId: string,
  departmentName: string,
): Member {
  const name = buildChineseName(index)
  return {
    id,
    name,
    phone: pickStatus(index) === 'pending' ? '' : buildPhone(index),
    email: buildEmail(name, index),
    departmentId,
    departmentName,
    status: pickStatus(index),
    roles: [ROLE_MEMBER],
    source: pickSource(index),
  }
}

export function buildMockMembers(): Member[] {
  const members: Member[] = [...ANCHOR_MEMBERS]
  let seq = 6

  for (const quota of LEAF_DEPT_QUOTAS) {
    const anchors = anchorsInDept(quota.departmentId)
    const generatedCount = quota.count - anchors.length
    for (let i = 0; i < generatedCount; i++) {
      members.push(buildGeneratedMember(`m-${seq}`, seq, quota.departmentId, quota.departmentName))
      seq++
    }
  }

  assignSpecialRoles(members)
  return members
}

function assignSpecialRoles(members: Member[]) {
  const orgAdmin = members.find((m) => m.id === 'm-10' && m.departmentId === 'dept-3')
  if (orgAdmin && !orgAdmin.roles.includes(ROLE_ORG_ADMIN)) {
    orgAdmin.roles = [...orgAdmin.roles, ROLE_ORG_ADMIN]
  }

  const budgetApprover = members.find((m) => m.departmentId === 'dept-6' && m.id !== 'm-2')
  if (budgetApprover && !budgetApprover.roles.includes(ROLE_BUDGET_APPROVER)) {
    budgetApprover.roles = [...budgetApprover.roles, ROLE_BUDGET_APPROVER]
  }

  const auditors = members
    .filter((m) => m.departmentId === 'dept-8' && m.status === 'active')
    .slice(0, 3)
  for (const auditor of auditors) {
    if (!auditor.roles.includes(ROLE_AUDITOR)) {
      auditor.roles = [...auditor.roles, ROLE_AUDITOR]
    }
  }

  const apiCallerCandidates = members
    .filter((m) => m.status === 'active' && !m.roles.includes(ROLE_SUPER_ADMIN))
    .slice(0, 50)
  for (const caller of apiCallerCandidates) {
    if (!caller.roles.includes(ROLE_API_CALLER)) {
      caller.roles = [...caller.roles, ROLE_API_CALLER]
    }
  }
}

export function countMembersByRole(members: Member[], roleName: string): number {
  return members.filter((m) => m.roles.includes(roleName)).length
}

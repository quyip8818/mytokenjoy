import type { Member } from '@/api/types'

export const mockMember: Member = {
  id: 'm-admin',
  companyId: 1,
  name: '管理员',
  phone: '13800000000',
  email: 'admin@test.com',
  departmentId: 'd1',
  departmentName: '总部',
  status: 'active',
  roles: ['超级管理员'],
  source: 'manual',
}

export const mockMembers: Member[] = [
  {
    id: 'm1',
    companyId: 1,
    name: '张三',
    phone: '13800000001',
    email: 'zhangsan@example.com',
    departmentId: 'd1',
    departmentName: '总部',
    status: 'active',
    roles: [],
    source: 'manual',
  },
  {
    id: 'm2',
    companyId: 1,
    name: '李四',
    phone: '13800000002',
    email: 'lisi@example.com',
    departmentId: 'd1',
    departmentName: '总部',
    status: 'active',
    roles: [],
    source: 'manual',
  },
]

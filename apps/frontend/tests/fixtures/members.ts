import type { Member } from '@/api/types'

export const mockMember: Member = {
  id: 'm-admin',
  companyId: '00000000-0000-7000-8000-000000000002',
  alias: '管理员',
  avatar: '',
  departmentId: 'd1',
  departmentName: '总部',
  status: 'active',
  roles: ['超级管理员'],
  source: 'manual',
}

export const mockMembers: Member[] = [
  {
    id: 'm1',
    companyId: '00000000-0000-7000-8000-000000000002',
    alias: '张三',
    avatar: '',
    departmentId: 'd1',
    departmentName: '总部',
    status: 'active',
    roles: [],
    source: 'manual',
  },
  {
    id: 'm2',
    companyId: '00000000-0000-7000-8000-000000000002',
    alias: '李四',
    avatar: '',
    departmentId: 'd1',
    departmentName: '总部',
    status: 'active',
    roles: [],
    source: 'manual',
  },
]

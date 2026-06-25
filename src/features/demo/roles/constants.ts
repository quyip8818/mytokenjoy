import type { Member } from '@/api/types'

export interface DemoSwitchableMember {
  id: string
  label: string
  displayName: string
  initials: string
  roleSummary: string
}

export const DEFAULT_DEMO_MEMBER_ID = 'm-admin'

export const DEMO_SWITCHABLE_MEMBERS: DemoSwitchableMember[] = [
  {
    id: 'm-admin',
    label: '管理员 · 超级管理员',
    displayName: '管理员',
    initials: '管',
    roleSummary: '超级管理员',
  },
  {
    id: 'm-2',
    label: '李四 · 组织管理员',
    displayName: '李四',
    initials: '李',
    roleSummary: '普通成员 · 组织管理员 · 预算审批员',
  },
  {
    id: 'm-1',
    label: '张三 · API 调用者',
    displayName: '张三',
    initials: '张',
    roleSummary: '普通成员 · API 调用者',
  },
  {
    id: 'm-auditor',
    label: '孙审计 · 只读审计员',
    displayName: '孙审计',
    initials: '孙',
    roleSummary: '普通成员 · 只读审计员',
  },
  {
    id: 'm-pure',
    label: '周八 · 普通成员',
    displayName: '周八',
    initials: '周',
    roleSummary: '普通成员',
  },
]

export function getSwitchableMember(memberId: string): DemoSwitchableMember | undefined {
  return DEMO_SWITCHABLE_MEMBERS.find((m) => m.id === memberId)
}

export function getMemberDisplay(member: Member): { displayName: string; initials: string } {
  const name = member.name.trim()
  return {
    displayName: name,
    initials: name.slice(0, 1) || '?',
  }
}

export function getDefaultHomePathForMemberId(memberId: string): string {
  const profile = getSwitchableMember(memberId)
  if (profile?.id === 'm-auditor') return '/audit/operations'
  if (profile?.id === 'm-1' || profile?.id === 'm-pure') return '/keys/mine'
  if (profile?.id === 'm-2') return '/keys/approval'
  return '/org/data-source'
}

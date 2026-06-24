export type DemoRole = 'admin' | 'tl' | 'member'

export const DEMO_ROLE_PROFILES = {
  admin: {
    label: '超管',
    memberId: 'm-admin',
    displayName: '管理员',
    initials: '管',
  },
  tl: {
    label: 'TL',
    memberId: 'm-2',
    displayName: '李四',
    initials: '李',
  },
  member: {
    label: '成员',
    memberId: 'm-1',
    displayName: '张三',
    initials: '张',
  },
} as const satisfies Record<
  DemoRole,
  { label: string; memberId: string; displayName: string; initials: string }
>

export const DEMO_ROLES: { id: DemoRole; label: string }[] = (
  Object.entries(DEMO_ROLE_PROFILES) as [DemoRole, (typeof DEMO_ROLE_PROFILES)[DemoRole]][]
).map(([id, profile]) => ({ id, label: profile.label }))

export const DEMO_ROLE_HOME: Record<DemoRole, string> = {
  admin: '/org/data-source',
  tl: '/keys/approval',
  member: '/keys/mine',
}

export function getDefaultHomePath(role: DemoRole): string {
  return DEMO_ROLE_HOME[role]
}

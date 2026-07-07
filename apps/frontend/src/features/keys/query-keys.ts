export const keysKeys = {
  all: ['keys'] as const,
  platform: () => [...keysKeys.all, 'platform'] as const,
  provider: () => [...keysKeys.all, 'provider'] as const,
  mine: (memberId: string) => [...keysKeys.all, 'mine', memberId] as const,
  quota: (memberId: string) => [...keysKeys.all, 'quota', memberId] as const,
  approvals: (tab: string, memberId?: string) =>
    [...keysKeys.all, 'approvals', tab, memberId ?? 'all'] as const,
}

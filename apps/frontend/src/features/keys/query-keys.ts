export const keysKeys = {
  all: ['keys'] as const,
  platformList: (departmentId?: string, type?: string) =>
    [...keysKeys.all, 'platform', departmentId ?? 'all', type ?? 'all'] as const,
  provider: () => [...keysKeys.all, 'provider'] as const,
  mine: (memberId: string) => [...keysKeys.all, 'mine', memberId] as const,
  budget: (memberId: string) => [...keysKeys.all, 'budget', memberId] as const,
}

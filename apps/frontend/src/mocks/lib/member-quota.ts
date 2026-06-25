import { mockPlatformKeys } from '../data'

// Personal quota pools and per-member key allocation for MSW handlers.

export interface MemberQuotaPool {
  personalQuota: number
}

export const mockMemberQuotaPools: Record<string, MemberQuotaPool> = {
  'm-admin': { personalQuota: 50000 },
  'm-1': { personalQuota: 10000 },
  'm-2': { personalQuota: 15000 },
  'm-4': { personalQuota: 12000 },
  'm-auditor': { personalQuota: 5000 },
  'm-pure': { personalQuota: 3000 },
}

const DEFAULT_PERSONAL_QUOTA = 5000

export function getPersonalQuota(memberId: string): number {
  return mockMemberQuotaPools[memberId]?.personalQuota ?? DEFAULT_PERSONAL_QUOTA
}

export function addPersonalQuota(memberId: string, amount: number): void {
  const current = getPersonalQuota(memberId)
  mockMemberQuotaPools[memberId] = { personalQuota: current + amount }
}

export function setPersonalQuota(memberId: string, personalQuota: number): void {
  mockMemberQuotaPools[memberId] = { personalQuota }
}

export function getAllocatedKeyQuota(memberId: string): number {
  return mockPlatformKeys
    .filter((k) => k.memberId === memberId && !k.budgetGroupId && k.status === 'active')
    .reduce((sum, k) => sum + k.quota, 0)
}

export function getUsedKeyQuota(memberId: string): number {
  return mockPlatformKeys
    .filter((k) => k.memberId === memberId && !k.budgetGroupId && k.status === 'active')
    .reduce((sum, k) => sum + k.used, 0)
}

export function getQuotaRemaining(memberId: string): number {
  return Math.max(0, getPersonalQuota(memberId) - getAllocatedKeyQuota(memberId))
}

export function buildQuotaSummary(memberId: string, reservedPool: number) {
  const totalQuota = getPersonalQuota(memberId)
  const used = getUsedKeyQuota(memberId)
  const remaining = getQuotaRemaining(memberId)
  return { totalQuota, used, remaining, reservedPool }
}

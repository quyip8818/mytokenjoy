import type { BudgetGroup } from '@/api/types'
import { mockPlatformKeys } from '../data'

// Budget group key allocation and validation for MSW handlers.

export function getAllocatedGroupKeyQuota(budgetGroupId: string): number {
  return mockPlatformKeys
    .filter((k) => k.budgetGroupId === budgetGroupId && k.status === 'active')
    .reduce((sum, k) => sum + k.quota, 0)
}

export function getGroupQuotaRemaining(group: BudgetGroup): number {
  const allocated = getAllocatedGroupKeyQuota(group.id)
  return Math.max(0, group.budget - group.consumed - allocated)
}

export function validateGroupKeyQuota(
  group: BudgetGroup,
  quota: number,
  excludeKeyId?: string,
): string | null {
  const allocated = mockPlatformKeys
    .filter(
      (k) =>
        k.budgetGroupId === group.id &&
        k.status === 'active' &&
        (!excludeKeyId || k.id !== excludeKeyId),
    )
    .reduce((sum, k) => sum + k.quota, 0)
  const remaining = group.budget - group.consumed - allocated
  if (quota > remaining) {
    return `预算组剩余可分配额度约 ¥${Math.max(0, remaining).toLocaleString()}`
  }
  return null
}

export function buildGroupQuotaSummary(group: BudgetGroup) {
  const allocated = getAllocatedGroupKeyQuota(group.id)
  const remaining = getGroupQuotaRemaining(group)
  return {
    totalQuota: group.budget,
    used: group.consumed,
    allocated,
    remaining,
  }
}

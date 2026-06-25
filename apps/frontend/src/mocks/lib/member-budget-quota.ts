import type { BudgetNode, Member, MemberBudgetQuota } from '@/api/types'
import { findBudgetNode, sumChildrenBudget } from '@/lib/budget'
import {
  getAllocatedKeyQuota,
  getPersonalQuota,
  getUsedKeyQuota,
  setPersonalQuota,
} from './member-quota'
import { findMemberById } from './query'

// Department member budget quota CRUD validation for MSW budget handlers.

export function getMemberQuotaCapacity(deptNode: BudgetNode): number {
  const reserved = deptNode.reservedPool ?? 0
  const childrenSum = sumChildrenBudget(deptNode)
  return Math.max(0, deptNode.budget - reserved - childrenSum)
}

export function buildMemberBudgetQuota(member: Member): MemberBudgetQuota {
  return {
    memberId: member.id,
    memberName: member.name,
    departmentId: member.departmentId,
    personalQuota: getPersonalQuota(member.id),
    allocated: getAllocatedKeyQuota(member.id),
    used: getUsedKeyQuota(member.id),
  }
}

export function validateMemberQuotaUpdate(
  budgetTree: BudgetNode[],
  members: Member[],
  memberId: string,
  personalQuota: number,
): string | null {
  const member = findMemberById(members, memberId)
  if (!member) return 'Member not found'

  const allocated = getAllocatedKeyQuota(memberId)
  if (personalQuota < allocated) {
    return `个人额度不能低于已分配 Key 额度（¥${allocated.toLocaleString()}）`
  }

  const deptNode = findBudgetNode(budgetTree, member.departmentId)
  if (!deptNode) return 'Department budget node not found'

  const capacity = getMemberQuotaCapacity(deptNode)
  const deptMembers = members.filter((m) => m.departmentId === member.departmentId)
  const otherSum = deptMembers
    .filter((m) => m.id !== memberId)
    .reduce((sum, m) => sum + getPersonalQuota(m.id), 0)

  if (otherSum + personalQuota > capacity) {
    return `超出部门可分配成员额度，当前剩余约 ¥${Math.max(0, capacity - otherSum).toLocaleString()}`
  }

  return null
}

export function applyMemberQuotaUpdate(
  members: Member[],
  memberId: string,
  personalQuota: number,
): MemberBudgetQuota {
  setPersonalQuota(memberId, personalQuota)
  const member = findMemberById(members, memberId)
  if (!member) {
    return {
      memberId,
      memberName: '',
      departmentId: '',
      personalQuota,
      allocated: getAllocatedKeyQuota(memberId),
      used: getUsedKeyQuota(memberId),
    }
  }
  return buildMemberBudgetQuota(member)
}

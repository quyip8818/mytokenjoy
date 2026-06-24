import type { BudgetNode, Member } from '@/api/types'
import { findBudgetNode } from '@/lib/budget'

export function getReservedPoolForDepartment(
  budgetTree: BudgetNode[],
  departmentId: string,
): number {
  const node = findBudgetNode(budgetTree, departmentId)
  return node?.reservedPool ?? 0
}

export function getReservedPoolForMember(
  budgetTree: BudgetNode[],
  members: Member[],
  memberId: string,
): number {
  const member = members.find((m) => m.id === memberId)
  if (!member) return 0
  return getReservedPoolForDepartment(budgetTree, member.departmentId)
}

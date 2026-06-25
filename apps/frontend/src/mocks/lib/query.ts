import type { Department, Member } from '@/api/types'
import { flattenDepartmentTree } from '@/lib/org'

export function collectDescendantDeptIds(departments: Department[], rootId: string): string[] {
  const flat = flattenDepartmentTree(departments)
  const childrenByParent = new Map<string, string[]>()
  for (const dept of flat) {
    if (dept.parentId) {
      const siblings = childrenByParent.get(dept.parentId) ?? []
      siblings.push(dept.id)
      childrenByParent.set(dept.parentId, siblings)
    }
  }

  const result: string[] = []
  const queue = [rootId]
  while (queue.length > 0) {
    const current = queue.shift()!
    result.push(current)
    const children = childrenByParent.get(current) ?? []
    queue.push(...children)
  }
  return result
}

export function filterMembersByDepartment(
  members: Member[],
  departments: Department[],
  departmentId: string,
  directOnly: boolean,
): Member[] {
  if (directOnly) {
    return members.filter((m) => m.departmentId === departmentId)
  }
  const allowed = new Set(collectDescendantDeptIds(departments, departmentId))
  return members.filter((m) => allowed.has(m.departmentId))
}

export function findMemberById(members: Member[], memberId: string): Member | undefined {
  return members.find((m) => m.id === memberId)
}

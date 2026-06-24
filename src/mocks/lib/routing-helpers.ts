import type { RoutingRule } from '@/api/types'
import { buildDeptParentMap } from '@/lib/org'
import { mockDepartments, mockModels, mockRoutingRules } from '../data'

export function deptParentMap() {
  return buildDeptParentMap(mockDepartments)
}

export function getRoutingRuleForDept(deptId: string): RoutingRule | undefined {
  let current: string | null | undefined = deptId
  const parents = deptParentMap()
  while (current) {
    const rule = mockRoutingRules.find((r) => r.nodeId === current)
    if (rule) return rule
    current = parents.get(current) ?? null
  }
  return undefined
}

export function getParentDeptId(deptId: string): string | null {
  return deptParentMap().get(deptId) ?? null
}

export function shrinkChildRoutingRules(parentNodeId: string, parentAllowed: string[]) {
  for (const rule of mockRoutingRules) {
    const parentId = getParentDeptId(rule.nodeId)
    if (parentId !== parentNodeId) continue
    rule.allowedModels = rule.allowedModels.filter((m) => parentAllowed.includes(m))
    if (rule.allowedModels.length === 0 && parentAllowed.length > 0) {
      rule.allowedModels = [...parentAllowed]
    }
    shrinkChildRoutingRules(rule.nodeId, rule.allowedModels)
  }
}

export { mockModels, mockRoutingRules }

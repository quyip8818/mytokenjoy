import type { AlertRule, BudgetGroup, BudgetNode, BudgetProjectView } from '@/api/types'

export interface AlertRuleView extends AlertRule {
  targetType: 'team' | 'project'
  targetId: string
  targetName: string
  departmentName?: string
}

export function isProjectNodeId(nodeId: string): boolean {
  return nodeId.startsWith('bg-')
}

export function alertRuleToView(rule: AlertRule, groups: BudgetGroup[]): AlertRuleView {
  const isProject = isProjectNodeId(rule.nodeId)
  const group = isProject ? groups.find((item) => item.id === rule.nodeId) : undefined
  return {
    ...rule,
    targetType: isProject ? 'project' : 'team',
    targetId: rule.nodeId,
    targetName: rule.nodeName,
    departmentName: group?.departmentIds[0],
  }
}

export function alertRuleFromView(
  view: Pick<
    AlertRuleView,
    'targetType' | 'targetId' | 'targetName' | 'thresholds' | 'notifyRoleIds' | 'enabled'
  >,
): Omit<AlertRule, 'id'> {
  return {
    nodeId: view.targetId,
    nodeName: view.targetName,
    thresholds: view.thresholds,
    notifyRoleIds: view.notifyRoleIds,
    enabled: view.enabled,
  }
}

export function thresholdClass(threshold: number): string {
  if (threshold >= 100) return 'bg-red-50 text-red-700 border-red-200'
  if (threshold >= 90) return 'bg-amber-50 text-amber-700 border-amber-200'
  return 'bg-emerald-50 text-emerald-700 border-emerald-200'
}

export function groupProjectsByTeam(
  projects: BudgetProjectView[],
  tree: BudgetNode[],
): { teamId: string; teamName: string; projects: BudgetProjectView[] }[] {
  const groups: { teamId: string; teamName: string; projects: BudgetProjectView[] }[] = []
  function walk(nodes: BudgetNode[]) {
    for (const node of nodes) {
      const nodeProjects = projects.filter((project) => project.departmentId === node.id)
      if (nodeProjects.length > 0) {
        groups.push({ teamId: node.id, teamName: node.name, projects: nodeProjects })
      }
      if (node.children) walk(node.children)
    }
  }
  walk(tree)
  return groups
}

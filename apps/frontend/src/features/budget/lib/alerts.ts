import type { AlertRule, Project, BudgetNode, ProjectView } from '@/api/types'

export interface AlertRuleView extends AlertRule {
  targetType: 'team' | 'project'
  targetId: string
  targetName: string
  departmentId?: string
}

export function isProjectNodeId(nodeId: string): boolean {
  return nodeId.startsWith('proj-')
}

export function alertRuleToView(rule: AlertRule, projects: Project[]): AlertRuleView {
  const isProject = isProjectNodeId(rule.nodeId)
  const project = isProject ? projects.find((item) => item.id === rule.nodeId) : undefined
  return {
    ...rule,
    targetType: isProject ? 'project' : 'team',
    targetId: rule.nodeId,
    targetName: rule.nodeName,
    departmentId: project?.ownerDepartmentId,
  }
}

export function alertRuleFromView(
  view: Pick<
    AlertRuleView,
    'targetType' | 'targetId' | 'targetName' | 'thresholds' | 'notifyRoleIds' | 'action' | 'enabled'
  >,
): Omit<AlertRule, 'id'> {
  return {
    nodeId: view.targetId,
    nodeName: view.targetName,
    thresholds: view.thresholds,
    notifyRoleIds: view.notifyRoleIds,
    action: view.action,
    enabled: view.enabled,
  }
}

export function thresholdClass(threshold: number): string {
  if (threshold >= 100) return 'bg-red-50 text-red-700 border-red-200'
  if (threshold >= 90) return 'bg-amber-50 text-amber-700 border-amber-200'
  return 'bg-emerald-50 text-emerald-700 border-emerald-200'
}

export function groupProjectsByTeam(
  projects: ProjectView[],
  tree: BudgetNode[],
): { teamId: string; teamName: string; projects: ProjectView[] }[] {
  const teams: { teamId: string; teamName: string; projects: ProjectView[] }[] = []
  function walk(nodes: BudgetNode[]) {
    for (const node of nodes) {
      const nodeProjects = projects.filter((project) => project.departmentId === node.id)
      if (nodeProjects.length > 0) {
        teams.push({ teamId: node.id, teamName: node.name, projects: nodeProjects })
      }
      if (node.children) walk(node.children)
    }
  }
  walk(tree)
  return teams
}

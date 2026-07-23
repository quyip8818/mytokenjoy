import type { Project, BudgetNode, OverrunPolicy } from '@/api/types'
import type { ProjectView } from '@/api/types'

export const DEFAULT_OVERRUN_POLICY: OverrunPolicy = 'hard_reject'

const OVERRUN_POLICY_LABELS: Record<OverrunPolicy, string> = {
  hard_reject: '硬拒绝',
  approval: '审批',
  downgrade: '降级',
}

export function formatOverrunPolicyLabel(policy: OverrunPolicy): string {
  return OVERRUN_POLICY_LABELS[policy]
}

export const BUDGET_WARNING_THRESHOLD = 70
export const BUDGET_DANGER_THRESHOLD = 90

export function getBudgetProgressTone(pct: number): 'danger' | 'warning' | 'default' | 'accent' {
  if (pct >= BUDGET_DANGER_THRESHOLD) return 'danger'
  if (pct >= BUDGET_WARNING_THRESHOLD) return 'warning'
  return 'default'
}

export function getBudgetProgressClass(pct: number, accent = false): string {
  const tone = getBudgetProgressTone(pct)
  if (tone === 'danger') return 'text-red-500'
  if (tone === 'warning') return 'text-amber-500'
  return accent ? 'text-blue-600' : 'text-muted-foreground'
}

export function sumChildrenBudget(node: BudgetNode): number {
  return (node.children ?? []).reduce((sum, child) => sum + child.budget, 0)
}

export function formatBudgetPeriodLabel(period: string | undefined): string {
  if (!period) return '-'
  const [year, month] = period.split('-')
  if (!year || !month) return period
  return `${year} 年 ${Number(month)} 月`
}

export function shiftBudgetPeriod(period: string, delta: number): string {
  if (!/^\d{4}-\d{2}$/.test(period)) {
    return period
  }
  const [y, m] = period.split('-').map(Number)
  const d = new Date(y, m - 1 + delta, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

export function computeUnallocated(node: BudgetNode): number {
  const reserved = node.reservedPool ?? 0
  const childrenSum = sumChildrenBudget(node)
  return Math.max(0, node.budget - node.consumed - reserved - childrenSum)
}

export function findBudgetNode(nodes: BudgetNode[], id: string): BudgetNode | null {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children) {
      const found = findBudgetNode(node.children, id)
      if (found) return found
    }
  }
  return null
}

export function updateBudgetNodeInTree(
  nodes: BudgetNode[],
  id: string,
  data: Partial<BudgetNode>,
): boolean {
  for (const node of nodes) {
    if (node.id === id) {
      Object.assign(node, data)
      return true
    }
    if (node.children && updateBudgetNodeInTree(node.children, id, data)) return true
  }
  return false
}

export function nodeReservedPool(node: BudgetNode): number {
  return node.reservedPool ?? 0
}

export function toProjectView(
  project: Project,
  departmentName: string,
  period: string,
  overrunPolicy: OverrunPolicy = DEFAULT_OVERRUN_POLICY,
): ProjectView {
  return {
    id: project.id,
    name: project.name,
    budget: project.budget,
    consumed: project.consumed,
    memberIds: project.memberIds,
    memberBudgets: project.memberBudgets,
    departmentId: project.ownerDepartmentId,
    departmentName,
    overrunPolicy,
    period,
    ownerId: project.ownerId,
  }
}

export function mapProjectsToViews(
  projects: Project[],
  nodeMap: Map<string, string>,
  period: string,
  overrunPolicy: OverrunPolicy = DEFAULT_OVERRUN_POLICY,
): ProjectView[] {
  return projects.map((project) => {
    const deptId = project.ownerDepartmentId
    const deptName = nodeMap.get(deptId) ?? ''
    return toProjectView(project, deptName, period, overrunPolicy)
  })
}

export function projectsForDepartment(projects: Project[], departmentId: string): Project[] {
  return projects.filter((project) => project.ownerDepartmentId === departmentId)
}

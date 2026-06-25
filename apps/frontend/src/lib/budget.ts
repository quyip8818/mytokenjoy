import type { BudgetNode } from '@/api/types'

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

export function computeUnallocated(node: BudgetNode): number {
  const reserved = node.reservedPool ?? 0
  const childrenSum = sumChildrenBudget(node)
  return Math.max(0, node.budget - node.consumed - reserved - childrenSum)
}

export function flattenBudgetNodes(nodes: BudgetNode[]): BudgetNode[] {
  const result: BudgetNode[] = []
  for (const node of nodes) {
    result.push(node)
    if (node.children) result.push(...flattenBudgetNodes(node.children))
  }
  return result
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

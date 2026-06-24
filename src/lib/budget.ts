import type { BudgetNode } from '@/api/types'

export function sumChildrenBudget(node: BudgetNode): number {
  return (node.children ?? []).reduce((sum, child) => sum + child.budget, 0)
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

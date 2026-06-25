import type { BudgetGroup, BudgetNode } from '@/api/types'

export interface BudgetWorkflowPayloads {
  'budget-node-edit': {
    node: BudgetNode
    parent?: BudgetNode | null
    onSuccess?: () => void
  }
  'budget-group-form': {
    group?: BudgetGroup
    tree?: BudgetNode[]
    onSuccess?: (id?: string) => void
  }
  'overrun-policy': {
    onSuccess?: () => void
  }
  'budget-impact-preview': {
    nodeId: string
    nodeName: string
    before: { budget: number; reservedPool: number }
    after: { budget: number; reservedPool: number }
    onSuccess?: () => void
  }
  'member-quota-config': {
    departmentId: string
    departmentName: string
    onSuccess?: () => void
  }
}

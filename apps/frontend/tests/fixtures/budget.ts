import type { BudgetGroup, BudgetNode } from '@/api/types'

export const mockBudgetGroups: BudgetGroup[] = [
  {
    id: 'bg1',
    name: '项目 A',
    budget: 10000,
    consumed: 2000,
    memberIds: ['m1'],
    departmentIds: ['d1'],
  },
]

export const mockBudgetTree: BudgetNode[] = [
  {
    id: 'n1',
    name: '总部',
    parentId: null,
    budget: 50000,
    consumed: 10000,
    period: '2026-01',
    children: [],
  },
]

import type { Project, BudgetNode } from '@/api/types'

export const mockProjects: Project[] = [
  {
    id: 'proj-1',
    name: '项目 A',
    budget: 10000,
    consumed: 2000,
    memberIds: ['m1'],
    ownerDepartmentId: 'd1',
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
    memberAvgBudget: 0,
    children: [],
  },
]

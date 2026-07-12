export const budgetKeys = {
  all: ['budget'] as const,
  tree: (period?: string) => [...budgetKeys.all, 'tree', period] as const,
  projects: () => [...budgetKeys.all, 'projects'] as const,
  overrunPolicy: () => [...budgetKeys.all, 'overrun-policy'] as const,
  alerts: () => [...budgetKeys.all, 'alerts'] as const,
  approvals: () => [...budgetKeys.all, 'approvals'] as const,
  memberBudgets: (departmentId: string) =>
    [...budgetKeys.all, 'member-budgets', departmentId] as const,
}

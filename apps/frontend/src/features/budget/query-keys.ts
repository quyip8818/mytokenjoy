export const budgetKeys = {
  all: ['budget'] as const,
  tree: (period?: string) => [...budgetKeys.all, 'tree', period] as const,
  groups: () => [...budgetKeys.all, 'groups'] as const,
  overrunPolicy: () => [...budgetKeys.all, 'overrun-policy'] as const,
  alerts: () => [...budgetKeys.all, 'alerts'] as const,
  approvals: () => [...budgetKeys.all, 'approvals'] as const,
  memberQuotas: (departmentId: string) =>
    [...budgetKeys.all, 'member-quotas', departmentId] as const,
}

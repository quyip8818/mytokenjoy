import type { CostQueryParams, UsageSeriesQuery } from '@/api/types'

export const queryKeys = {
  session: {
    all: ['session'] as const,
    current: () => [...queryKeys.session.all, 'current'] as const,
  },
  budget: {
    all: ['budget'] as const,
    tree: () => [...queryKeys.budget.all, 'tree'] as const,
    groups: () => [...queryKeys.budget.all, 'groups'] as const,
    overrunPolicy: () => [...queryKeys.budget.all, 'overrun-policy'] as const,
    memberQuotas: (departmentId: string) =>
      [...queryKeys.budget.all, 'member-quotas', departmentId] as const,
  },
  org: {
    all: ['org'] as const,
    departmentTree: () => [...queryKeys.org.all, 'department-tree'] as const,
    members: (params: unknown) => [...queryKeys.org.all, 'members', params] as const,
    rolesInit: () => [...queryKeys.org.all, 'roles-init'] as const,
    roleMembers: (roleId: string) => [...queryKeys.org.all, 'role-members', roleId] as const,
    dataSource: () => [...queryKeys.org.all, 'data-source'] as const,
    syncLogs: () => [...queryKeys.org.all, 'sync-logs'] as const,
    approvalPendingCount: () => [...queryKeys.org.all, 'approval-pending-count'] as const,
  },
  keys: {
    all: ['keys'] as const,
    platform: () => [...queryKeys.keys.all, 'platform'] as const,
    provider: () => [...queryKeys.keys.all, 'provider'] as const,
    mine: (memberId: string) => [...queryKeys.keys.all, 'mine', memberId] as const,
    quota: (memberId: string) => [...queryKeys.keys.all, 'quota', memberId] as const,
    approvals: (tab: string, memberId?: string) =>
      [...queryKeys.keys.all, 'approvals', tab, memberId ?? 'all'] as const,
  },
  models: {
    all: ['models'] as const,
    list: () => [...queryKeys.models.all, 'list'] as const,
    routing: () => [...queryKeys.models.all, 'routing'] as const,
  },
  audit: {
    all: ['audit'] as const,
    settings: () => [...queryKeys.audit.all, 'settings'] as const,
    members: () => [...queryKeys.audit.all, 'members'] as const,
    operations: (filter: unknown) => [...queryKeys.audit.all, 'operations', filter] as const,
    calls: (filter: unknown) => [...queryKeys.audit.all, 'calls', filter] as const,
  },
  dashboard: {
    all: ['dashboard'] as const,
    cost: (query: CostQueryParams, drill: unknown, granularity: string) =>
      [...queryKeys.dashboard.all, 'cost', query, drill, granularity] as const,
    usage: () => [...queryKeys.dashboard.all, 'usage'] as const,
    usageSeries: (query: UsageSeriesQuery) =>
      [...queryKeys.dashboard.all, 'usage-series', query] as const,
  },
}

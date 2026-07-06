import type { CostQueryParams, UsageSeriesQuery } from '@/api/types'

export const queryKeys = {
  session: {
    all: ['session'] as const,
    current: () => [...queryKeys.session.all, 'current'] as const,
  },
  budget: {
    all: ['budget'] as const,
    tree: (period?: string) => [...queryKeys.budget.all, 'tree', period] as const,
    groups: () => [...queryKeys.budget.all, 'groups'] as const,
    overrunPolicy: () => [...queryKeys.budget.all, 'overrun-policy'] as const,
    alerts: () => [...queryKeys.budget.all, 'alerts'] as const,
    approvals: () => [...queryKeys.budget.all, 'approvals'] as const,
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
    operations: (params: unknown) => [...queryKeys.audit.all, 'operations', params] as const,
    calls: (params: unknown) => [...queryKeys.audit.all, 'calls', params] as const,
    models: () => [...queryKeys.audit.all, 'models'] as const,
  },
  billing: {
    all: ['billing'] as const,
    wallet: () => [...queryKeys.billing.all, 'wallet'] as const,
    rechargeRecords: () => [...queryKeys.billing.all, 'recharge-records'] as const,
  },
  me: {
    all: ['me'] as const,
    dashboard: () => [...queryKeys.me.all, 'dashboard'] as const,
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

export const orgKeys = {
  all: ['org'] as const,
  tree: () => [...orgKeys.all, 'tree'] as const,
  departmentTree: () => [...orgKeys.all, 'department-tree'] as const,
  roles: () => [...orgKeys.all, 'roles'] as const,
  permissions: () => [...orgKeys.all, 'permissions'] as const,
  members: (params: unknown) => [...orgKeys.all, 'members', params] as const,
  roleMembers: (roleId: string) => [...orgKeys.all, 'role-members', roleId] as const,
  dataSource: () => [...orgKeys.all, 'data-source'] as const,
  syncLogs: () => [...orgKeys.all, 'sync-logs'] as const,
  approvalPendingCount: () => [...orgKeys.all, 'approval-pending-count'] as const,
}

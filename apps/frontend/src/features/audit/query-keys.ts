export const auditKeys = {
  all: ['audit'] as const,
  settings: () => [...auditKeys.all, 'settings'] as const,
  members: () => [...auditKeys.all, 'members'] as const,
  operations: (params: unknown) => [...auditKeys.all, 'operations', params] as const,
  calls: (params: unknown) => [...auditKeys.all, 'calls', params] as const,
  models: () => [...auditKeys.all, 'models'] as const,
}

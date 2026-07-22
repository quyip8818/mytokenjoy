export const approvalKeys = {
  all: ['approval'] as const,
  list: (status?: string, type?: string) =>
    [...approvalKeys.all, 'list', status ?? 'all', type ?? 'all'] as const,
  detail: (id: string) => [...approvalKeys.all, 'detail', id] as const,
  preCheck: (id: string) => [...approvalKeys.all, 'preCheck', id] as const,
  pendingCount: () => [...approvalKeys.all, 'pending-count'] as const,
}

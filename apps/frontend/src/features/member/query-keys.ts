export const memberKeys = {
  all: ['me'] as const,
  dashboard: () => [...memberKeys.all, 'dashboard'] as const,
  callLogs: (params: unknown) => [...memberKeys.all, 'call-logs', params] as const,
}

export const memberKeys = {
  all: ['member'] as const,
  dashboard: () => [...memberKeys.all, 'dashboard'] as const,
  callLogs: (params: unknown) => [...memberKeys.all, 'call-logs', params] as const,
}

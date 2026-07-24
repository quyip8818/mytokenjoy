export const mydashboardKeys = {
  all: ['mydashboard'] as const,
  dashboard: () => [...mydashboardKeys.all, 'dashboard'] as const,
  callLogs: (params: unknown) => [...mydashboardKeys.all, 'call-logs', params] as const,
}

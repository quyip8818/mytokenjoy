import type { CostQueryParams } from '@/api/types'

export const dashboardKeys = {
  all: ['dashboard'] as const,
  cost: (query: CostQueryParams, drill: unknown, granularity: string) =>
    [...dashboardKeys.all, 'cost', query, drill, granularity] as const,
  usage: () => [...dashboardKeys.all, 'usage'] as const,
}

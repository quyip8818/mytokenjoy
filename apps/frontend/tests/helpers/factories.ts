import { vi } from 'vitest'
import type { Paginated } from '@/api/types'
import type { AppApis } from '@/api/app-apis'

/**
 * 创建一个标准分页响应
 */
export function createPaginatedResponse<T>(
  items: T[],
  opts?: { page?: number; pageSize?: number; total?: number },
): Paginated<T> {
  return {
    items,
    total: opts?.total ?? items.length,
    page: opts?.page ?? 1,
    pageSize: opts?.pageSize ?? 20,
  }
}

/**
 * 创建 dashboardApi mock，所有方法默认返回空数据
 */
export function createDashboardApiMock(
  overrides?: Partial<AppApis['dashboardApi']>,
): AppApis['dashboardApi'] {
  return {
    getCostSummary: vi.fn().mockResolvedValue({
      totalCost: 0,
      totalCostMom: 0,
      totalTokens: 0,
      totalRequests: 0,
      totalRequestsMom: 0,
      avgCostPerRequest: 0,
      avgCostPerRequestMom: 0,
      avgCostPerMember: 0,
      avgCostPerMemberMom: 0,
    }),
    getDailyCosts: vi.fn().mockResolvedValue([]),
    getDepartmentCosts: vi.fn().mockResolvedValue([]),
    getDepartmentMemberCosts: vi.fn().mockResolvedValue([]),
    getTopConsumers: vi.fn().mockResolvedValue([]),
    getModelUsage: vi.fn().mockResolvedValue([]),
    getDepartmentUsage: vi.fn().mockResolvedValue([]),
    ...overrides,
  }
}

/**
 * 创建 auditApi mock，所有方法默认返回空数据
 */
export function createAuditApiMock(
  overrides?: Partial<AppApis['auditApi']>,
): AppApis['auditApi'] {
  return {
    getOperations: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 20 }),
    getOperationsTimeline: vi.fn().mockResolvedValue([]),
    getCalls: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 20 }),
    getCallsSummary: vi.fn().mockResolvedValue({
      totalCalls: 0,
      errorCount: 0,
      errorRate: 0,
      avgLatencyMs: 0,
    }),
    getSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
    updateSettings: vi.fn().mockResolvedValue({ contentRetentionEnabled: true }),
    ...overrides,
  }
}

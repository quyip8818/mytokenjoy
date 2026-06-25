import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import type { CostPeriod } from '@/api/types'
import {
  buildCostSummary,
  buildDailyCosts,
  getDepartmentCostsForParent,
  getDepartmentMemberCosts,
  getTopConsumers,
  mockModelUsage,
  mockTeamUsage,
} from '../fixtures/dashboard'

function parsePeriod(url: URL): CostPeriod {
  const period = url.searchParams.get('period') as CostPeriod | null
  if (period === 'last_month' || period === 'last_7_days') return period
  return 'current_month'
}

export const dashboardHandlers = [
  http.get(`${API_BASE_PATH}/dashboard/cost/summary`, ({ request }) => {
    const period = parsePeriod(new URL(request.url))
    return HttpResponse.json(buildCostSummary(period))
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/departments`, ({ request }) => {
    const url = new URL(request.url)
    const period = parsePeriod(url)
    const parentId = url.searchParams.get('parentId')
    return HttpResponse.json(getDepartmentCostsForParent(parentId, period))
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/departments/:deptId/members`, ({ request, params }) => {
    const period = parsePeriod(new URL(request.url))
    return HttpResponse.json(getDepartmentMemberCosts(params.deptId as string, period))
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/daily`, ({ request }) => {
    const period = parsePeriod(new URL(request.url))
    return HttpResponse.json(buildDailyCosts(period))
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/top`, ({ request }) => {
    const url = new URL(request.url)
    const period = parsePeriod(url)
    const limit = Number(url.searchParams.get('limit') ?? 5)
    return HttpResponse.json(getTopConsumers(limit, period))
  }),
  http.get(`${API_BASE_PATH}/dashboard/usage/models`, () => {
    return HttpResponse.json(mockModelUsage)
  }),
  http.get(`${API_BASE_PATH}/dashboard/usage/teams`, () => {
    return HttpResponse.json(mockTeamUsage)
  }),
]

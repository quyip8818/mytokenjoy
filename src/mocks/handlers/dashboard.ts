import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import {
  mockCostSummary,
  mockDepartmentCosts,
  mockDailyCosts,
  mockTopConsumers,
  mockModelUsage,
  mockTeamUsage,
} from '../data'

export const dashboardHandlers = [
  // ========== 数据看板 ==========
  http.get(`${API_BASE_PATH}/dashboard/cost/summary`, () => {
    return HttpResponse.json(mockCostSummary)
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/departments`, () => {
    return HttpResponse.json(mockDepartmentCosts)
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/daily`, () => {
    return HttpResponse.json(mockDailyCosts)
  }),
  http.get(`${API_BASE_PATH}/dashboard/cost/top`, () => {
    return HttpResponse.json(mockTopConsumers)
  }),
  http.get(`${API_BASE_PATH}/dashboard/usage/models`, () => {
    return HttpResponse.json(mockModelUsage)
  }),
  http.get(`${API_BASE_PATH}/dashboard/usage/teams`, () => {
    return HttpResponse.json(mockTeamUsage)
  }),
]

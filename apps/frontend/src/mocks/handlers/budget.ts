import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { findBudgetNode, updateBudgetNodeInTree } from '@/lib/budget'
import {
  applyMemberQuotaUpdate,
  buildMemberBudgetQuota,
  validateMemberQuotaUpdate,
} from '../lib/member-budget-quota'
import {
  mockBudgetTree,
  mockBudgetGroups,
  mockAlertRules,
  mockOverrunPolicy,
  mockMembers,
} from '../data'

export const budgetHandlers = [
  // ========== 预算管理 ==========
  http.get(`${API_BASE_PATH}/budget/tree`, () => {
    return HttpResponse.json(mockBudgetTree)
  }),
  http.put(`${API_BASE_PATH}/budget/nodes/:id`, async ({ params, request }) => {
    await delay(300)
    const body = (await request.json()) as { budget: number; reservedPool?: number }
    const id = params.id as string
    const existing = findBudgetNode(mockBudgetTree, id)
    if (!existing) {
      return HttpResponse.json({ message: 'Node not found' }, { status: 404 })
    }
    updateBudgetNodeInTree(mockBudgetTree, id, {
      budget: body.budget,
      reservedPool: body.reservedPool ?? existing.reservedPool ?? 0,
    })
    return HttpResponse.json(findBudgetNode(mockBudgetTree, id))
  }),
  http.get(`${API_BASE_PATH}/budget/departments/:deptId/member-quotas`, ({ params }) => {
    const deptId = params.deptId as string
    const deptNode = findBudgetNode(mockBudgetTree, deptId)
    if (!deptNode) {
      return HttpResponse.json({ message: 'Department not found' }, { status: 404 })
    }
    const quotas = mockMembers
      .filter((m) => m.departmentId === deptId)
      .map((m) => buildMemberBudgetQuota(m))
    return HttpResponse.json(quotas)
  }),
  http.put(`${API_BASE_PATH}/budget/members/:memberId`, async ({ params, request }) => {
    await delay(300)
    const memberId = params.memberId as string
    const body = (await request.json()) as { personalQuota: number }
    const error = validateMemberQuotaUpdate(
      mockBudgetTree,
      mockMembers,
      memberId,
      body.personalQuota,
    )
    if (error) {
      return HttpResponse.json({ message: error }, { status: 422 })
    }
    return HttpResponse.json(applyMemberQuotaUpdate(mockMembers, memberId, body.personalQuota))
  }),
  http.get(`${API_BASE_PATH}/budget/groups`, () => {
    return HttpResponse.json(mockBudgetGroups)
  }),
  http.post(`${API_BASE_PATH}/budget/groups`, async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as Record<string, unknown>
    const group = { id: `bg-${Date.now()}`, consumed: 0, ...body } as (typeof mockBudgetGroups)[0]
    mockBudgetGroups.push(group)
    return HttpResponse.json(group)
  }),
  http.put(`${API_BASE_PATH}/budget/groups/:id`, async ({ params, request }) => {
    await delay(300)
    const body = (await request.json()) as Partial<(typeof mockBudgetGroups)[0]>
    const idx = mockBudgetGroups.findIndex((g) => g.id === params.id)
    if (idx < 0) return HttpResponse.json({ message: 'Not found' }, { status: 404 })
    mockBudgetGroups[idx] = { ...mockBudgetGroups[idx], ...body }
    return HttpResponse.json(mockBudgetGroups[idx])
  }),
  http.delete(`${API_BASE_PATH}/budget/groups/:id`, ({ params }) => {
    const idx = mockBudgetGroups.findIndex((g) => g.id === params.id)
    if (idx >= 0) mockBudgetGroups.splice(idx, 1)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/budget/overrun-policy`, () => {
    return HttpResponse.json(mockOverrunPolicy)
  }),
  http.put(`${API_BASE_PATH}/budget/overrun-policy`, async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as typeof mockOverrunPolicy
    Object.assign(mockOverrunPolicy, body)
    return HttpResponse.json(mockOverrunPolicy)
  }),
  http.get(`${API_BASE_PATH}/budget/alerts`, () => {
    return HttpResponse.json(mockAlertRules)
  }),
  http.post(`${API_BASE_PATH}/budget/alerts`, async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ id: `alert-${Date.now()}`, ...body })
  }),
  http.put(`${API_BASE_PATH}/budget/alerts/:id`, async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(body)
  }),
  http.delete(`${API_BASE_PATH}/budget/alerts/:id`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
]

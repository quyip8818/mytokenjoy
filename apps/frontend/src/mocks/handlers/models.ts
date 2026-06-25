import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import type { RoutingRule } from '@/api/types'
import {
  getRoutingRuleForDept,
  getParentDeptId,
  shrinkChildRoutingRules,
  mockModels,
  mockRoutingRules,
} from '../lib/routing-helpers'

export const modelsHandlers = [
  // ========== 模型路由 ==========
  http.get(`${API_BASE_PATH}/models`, () => {
    return HttpResponse.json(mockModels)
  }),
  http.post(`${API_BASE_PATH}/models`, async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as {
      name: string
      displayName: string
      inputPrice: number
      outputPrice: number
    }
    const model = {
      id: `model-${Date.now()}`,
      provider: 'custom' as const,
      name: body.name,
      displayName: body.displayName || body.name,
      inputPrice: body.inputPrice,
      outputPrice: body.outputPrice,
      maxContext: 128000,
      enabled: true,
      capabilities: ['chat'],
    }
    mockModels.push(model)
    return HttpResponse.json(model)
  }),
  http.put(`${API_BASE_PATH}/models/:id/toggle`, async ({ params, request }) => {
    await delay(300)
    const body = (await request.json()) as { enabled: boolean }
    const model = mockModels.find((m) => m.id === params.id)
    if (model) model.enabled = body.enabled
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/models/routing`, () => {
    return HttpResponse.json(mockRoutingRules)
  }),
  http.get(`${API_BASE_PATH}/models/routing/resolve`, ({ request }) => {
    const url = new URL(request.url)
    const deptId = url.searchParams.get('deptId') ?? ''
    const rule = getRoutingRuleForDept(deptId)
    if (!rule) {
      return HttpResponse.json({
        inherited: false,
        allowedModels: mockModels.filter((m) => m.enabled).map((m) => m.name),
        parentCount: mockModels.length,
      })
    }
    const parentId = getParentDeptId(rule.nodeId)
    const parentRule = parentId ? mockRoutingRules.find((r) => r.nodeId === parentId) : undefined
    const parentCount = parentRule?.allowedModels.length ?? rule.allowedModels.length
    let allowedModels = rule.allowedModels
    if (rule.inherited && parentRule) {
      allowedModels = rule.allowedModels.filter((m) => parentRule.allowedModels.includes(m))
      if (allowedModels.length === 0) allowedModels = [...parentRule.allowedModels]
    }
    return HttpResponse.json({
      inherited: rule.inherited,
      allowedModels,
      parentCount,
    })
  }),
  http.put(`${API_BASE_PATH}/models/routing/:id`, async ({ params, request }) => {
    await delay(300)
    const body = (await request.json()) as Partial<RoutingRule>
    const idx = mockRoutingRules.findIndex((r) => r.id === params.id)
    if (idx < 0) return HttpResponse.json({ message: 'Not found' }, { status: 404 })
    const prev = mockRoutingRules[idx]
    mockRoutingRules[idx] = { ...prev, ...body }
    const updated = mockRoutingRules[idx]
    if (body.allowedModels) {
      shrinkChildRoutingRules(updated.nodeId, updated.allowedModels)
    }
    return HttpResponse.json(updated)
  }),
]

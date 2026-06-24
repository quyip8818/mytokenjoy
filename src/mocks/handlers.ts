import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import type { RoutingRule, BatchImportRow } from '../api/types'
import { findBudgetNode, updateBudgetNodeInTree } from '@/lib/budget'
import { flattenDepartmentTree, buildDeptParentMap } from '@/lib/org'
import { getReservedPoolForMember } from './lib/budget-lookup'
import { paginate } from './lib/paginate'
import { filterMembersByDepartment, findMemberById } from './lib/query'
import {
  mockDataSourceStatus,
  mockSyncConfig,
  mockSyncLogs,
  mockImportFailures,
  mockDepartments,
  mockMembers,
  mockRoles,
  mockPermissions,
  mockBudgetTree,
  mockBudgetGroups,
  mockAlertRules,
  mockOverrunPolicy,
  mockProviderKeys,
  mockPlatformKeys,
  mockApprovals,
  mockModels,
  mockRoutingRules,
  mockCostSummary,
  mockDepartmentCosts,
  mockDailyCosts,
  mockTopConsumers,
  mockModelUsage,
  mockTeamUsage,
  mockOperationLogs,
  mockCallLogs,
} from './data'

function parseIntParam(value: string | null, fallback: number): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

const deptParentMap = () => buildDeptParentMap(mockDepartments)

function getRoutingRuleForDept(deptId: string): RoutingRule | undefined {
  let current: string | null | undefined = deptId
  const parents = deptParentMap()
  while (current) {
    const rule = mockRoutingRules.find((r) => r.nodeId === current)
    if (rule) return rule
    current = parents.get(current) ?? null
  }
  return undefined
}

function getParentDeptId(deptId: string): string | null {
  return deptParentMap().get(deptId) ?? null
}

function shrinkChildRoutingRules(parentNodeId: string, parentAllowed: string[]) {
  for (const rule of mockRoutingRules) {
    const parentId = getParentDeptId(rule.nodeId)
    if (parentId !== parentNodeId) continue
    rule.allowedModels = rule.allowedModels.filter((m) => parentAllowed.includes(m))
    if (rule.allowedModels.length === 0 && parentAllowed.length > 0) {
      rule.allowedModels = [...parentAllowed]
    }
    shrinkChildRoutingRules(rule.nodeId, rule.allowedModels)
  }
}

export const handlers = [
  // ========== 数据源 ==========
  http.get(`${API_BASE_PATH}/org/data-source/status`, () => {
    return HttpResponse.json(mockDataSourceStatus)
  }),
  http.post(`${API_BASE_PATH}/org/data-source/test`, async () => {
    await delay(1000)
    return HttpResponse.json({ success: true })
  }),
  http.put(`${API_BASE_PATH}/org/data-source`, () => {
    mockDataSourceStatus.connected = true
    mockDataSourceStatus.platform = 'feishu'
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/org/data-source/search`, ({ request }) => {
    const url = new URL(request.url)
    const keyword = url.searchParams.get('keyword') || ''
    if (!keyword) return HttpResponse.json({ name: '', department: '', mappingOk: false })
    return HttpResponse.json({
      name: keyword === '张三' ? '张三' : `${keyword}（模拟）`,
      department: '技术部 > 后端组',
      mappingOk: true,
    })
  }),
  http.post(`${API_BASE_PATH}/org/data-source/import`, async () => {
    await delay(2000)
    return HttpResponse.json({
      successMembers: 120,
      successDepartments: 5,
      failures: mockImportFailures,
    })
  }),
  http.post(`${API_BASE_PATH}/org/data-source/import/retry`, async () => {
    await delay(500)
    return HttpResponse.json({ successMembers: 1, successDepartments: 0, failures: [] })
  }),

  // ========== 同步 ==========
  http.get(`${API_BASE_PATH}/org/sync/config`, () => {
    return HttpResponse.json(mockSyncConfig)
  }),
  http.put(`${API_BASE_PATH}/org/sync/config`, async ({ request }) => {
    const body = (await request.json()) as typeof mockSyncConfig
    Object.assign(mockSyncConfig, body)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post(`${API_BASE_PATH}/org/sync/trigger`, async () => {
    await delay(1500)
    return HttpResponse.json({ successMembers: 3, successDepartments: 0, failures: [] })
  }),
  http.get(`${API_BASE_PATH}/org/sync/logs`, ({ request }) => {
    const url = new URL(request.url)
    const page = parseIntParam(url.searchParams.get('page'), 1)
    const pageSize = parseIntParam(url.searchParams.get('pageSize'), 10)
    return HttpResponse.json(paginate(mockSyncLogs, page, pageSize))
  }),

  // ========== 部门 ==========
  http.get(`${API_BASE_PATH}/org/departments/tree`, () => {
    return HttpResponse.json(mockDepartments)
  }),
  http.post(`${API_BASE_PATH}/org/departments`, async ({ request }) => {
    const body = (await request.json()) as { name: string; parentId: string }
    return HttpResponse.json({
      id: `dept-${Date.now()}`,
      name: body.name,
      parentId: body.parentId,
      memberCount: 0,
    })
  }),
  http.put(`${API_BASE_PATH}/org/departments/:id`, async ({ request }) => {
    const body = (await request.json()) as { name: string }
    return HttpResponse.json({ id: 'dept-x', name: body.name, parentId: null, memberCount: 0 })
  }),
  http.delete(`${API_BASE_PATH}/org/departments/:id`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),

  // ========== 成员 ==========
  http.get(`${API_BASE_PATH}/org/members`, ({ request }) => {
    const url = new URL(request.url)
    const deptId = url.searchParams.get('departmentId')
    const keyword = url.searchParams.get('keyword')
    const directOnly = url.searchParams.get('directOnly') === 'true'
    const page = parseIntParam(url.searchParams.get('page'), 1)
    const pageSize = parseIntParam(url.searchParams.get('pageSize'), 20)
    let items = [...mockMembers]
    if (deptId) {
      items = filterMembersByDepartment(mockMembers, mockDepartments, deptId, directOnly)
    }
    if (keyword) items = items.filter((m) => m.name.includes(keyword))
    return HttpResponse.json(paginate(items, page, pageSize))
  }),
  http.post(`${API_BASE_PATH}/org/members`, async ({ request }) => {
    const body = (await request.json()) as Record<string, string>
    return HttpResponse.json({
      id: `m-${Date.now()}`,
      ...body,
      status: 'active',
      roles: ['普通成员'],
      source: 'manual',
    })
  }),
  http.put(`${API_BASE_PATH}/org/members/:id`, async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(body)
  }),
  http.delete(`${API_BASE_PATH}/org/members`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.put(`${API_BASE_PATH}/org/members/status`, async ({ request }) => {
    const body = (await request.json()) as { ids: string[]; status: 'active' | 'inactive' }
    for (const id of body.ids) {
      const member = mockMembers.find((m) => m.id === id)
      if (member) member.status = body.status
      if (body.status === 'inactive') {
        mockPlatformKeys.forEach((k) => {
          if (k.memberId === id) k.status = 'disabled'
        })
      }
    }
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post(`${API_BASE_PATH}/org/members/transfer`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post(`${API_BASE_PATH}/org/members/invite`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post(`${API_BASE_PATH}/org/members/batch-invite`, async ({ request }) => {
    const body = (await request.json()) as { ids?: string[] }
    const targets = body.ids?.length
      ? mockMembers.filter((m) => body.ids!.includes(m.id))
      : mockMembers.filter((m) => m.status === 'pending' || m.status === 'inactive')
    return HttpResponse.json({ sent: targets.length })
  }),
  http.post(`${API_BASE_PATH}/org/members/batch-import`, async ({ request }) => {
    const body = (await request.json()) as { rows: BatchImportRow[] }
    const failures: { row: number; reason: string }[] = []
    let imported = 0
    const flat = flattenDepartmentTree(mockDepartments)
    body.rows.forEach((row, index) => {
      const dept = flat.find((d) => d.name === row.departmentName)
      if (!dept) {
        failures.push({ row: index + 1, reason: 'Department not found' })
        return
      }
      mockMembers.push({
        id: `m-import-${Date.now()}-${index}`,
        name: row.name,
        phone: row.phone,
        email: row.email,
        departmentId: dept.id,
        departmentName: dept.name,
        status: 'active',
        roles: ['普通成员'],
        source: 'imported',
      })
      imported++
    })
    return HttpResponse.json({ imported, failures })
  }),

  // ========== 角色 ==========
  http.get(`${API_BASE_PATH}/org/roles`, () => {
    return HttpResponse.json(mockRoles)
  }),
  http.post(`${API_BASE_PATH}/org/roles`, async ({ request }) => {
    const body = (await request.json()) as { name: string; permissions: string[] }
    return HttpResponse.json({ id: `role-${Date.now()}`, ...body, type: 'custom', memberCount: 0 })
  }),
  http.put(`${API_BASE_PATH}/org/roles/:id`, async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(body)
  }),
  http.delete(`${API_BASE_PATH}/org/roles/:id`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/org/roles/:roleId/members`, ({ params }) => {
    const role = mockRoles.find((r) => r.id === params.roleId)
    if (!role) return HttpResponse.json([])
    return HttpResponse.json(mockMembers.filter((m) => m.roles.includes(role.name)))
  }),
  http.post(`${API_BASE_PATH}/org/roles/:roleId/members`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.delete(`${API_BASE_PATH}/org/roles/:roleId/members/:memberId`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/org/permissions`, () => {
    return HttpResponse.json(mockPermissions)
  }),

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

  // ========== API-KEY 管理 ==========
  http.get(`${API_BASE_PATH}/keys/provider`, () => {
    return HttpResponse.json(mockProviderKeys)
  }),
  http.post(`${API_BASE_PATH}/keys/provider`, async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({
      id: `pk-${Date.now()}`,
      ...body,
      keyPrefix: 'sk-new...',
      status: 'active',
      balance: null,
      lastUsed: null,
      createdAt: '2026-06-19',
      rotateEnabled: false,
    })
  }),
  http.put(`${API_BASE_PATH}/keys/provider/:id/toggle`, async () => {
    await delay(300)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post(`${API_BASE_PATH}/keys/provider/:id/rotate`, async ({ params }) => {
    await delay(1000)
    const idx = mockProviderKeys.findIndex((k) => k.id === params.id)
    if (idx === -1) {
      return HttpResponse.json({ message: 'Not found' }, { status: 404 })
    }
    const updated = {
      ...mockProviderKeys[idx],
      keyPrefix: `sk-rot-${Date.now().toString(36)}...`,
      lastUsed: new Date().toISOString().slice(0, 16).replace('T', ' '),
    }
    mockProviderKeys[idx] = updated
    return HttpResponse.json(updated)
  }),
  http.delete(`${API_BASE_PATH}/keys/provider/:id`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/keys/platform`, ({ request }) => {
    const url = new URL(request.url)
    const memberId = url.searchParams.get('memberId')
    let items = [...mockPlatformKeys]
    if (memberId) {
      items = items.filter((k) => k.memberId === memberId)
    }
    return HttpResponse.json({
      items,
      total: items.length,
      page: 1,
      pageSize: 20,
    })
  }),
  http.get(`${API_BASE_PATH}/keys/platform/quota-summary`, ({ request }) => {
    const url = new URL(request.url)
    const memberId = url.searchParams.get('memberId') ?? 'm-1'
    const keys = mockPlatformKeys.filter((k) => k.memberId === memberId && k.status === 'active')
    const totalQuota = keys.reduce((sum, k) => sum + k.quota, 0)
    const used = keys.reduce((sum, k) => sum + k.used, 0)
    return HttpResponse.json({
      totalQuota,
      used,
      remaining: Math.max(0, totalQuota - used),
      reservedPool: getReservedPoolForMember(mockBudgetTree, mockMembers, memberId),
    })
  }),
  http.post(`${API_BASE_PATH}/keys/platform`, async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as Record<string, unknown>
    const fullKey = `tj-${Date.now()}-demo-secret-key`
    const memberId = (body.memberId as string) ?? null
    const member = memberId ? findMemberById(mockMembers, memberId) : undefined
    const newKey = {
      id: `plk-${Date.now()}`,
      name: body.name as string,
      keyPrefix: `${fullKey.slice(0, 12)}...`,
      fullKey,
      memberId,
      memberName: member?.name ?? null,
      appName: (body.appName as string) ?? null,
      quota: body.quota as number,
      used: 0,
      modelWhitelist: body.modelWhitelist as string[],
      status: 'active' as const,
      createdAt: new Date().toISOString().slice(0, 10),
      expiresAt: null,
    }
    mockPlatformKeys.push(newKey)
    return HttpResponse.json(newKey)
  }),
  http.put(`${API_BASE_PATH}/keys/platform/:id`, async ({ params, request }) => {
    await delay(300)
    const body = (await request.json()) as Record<string, unknown>
    const idx = mockPlatformKeys.findIndex((k) => k.id === params.id)
    if (idx >= 0) {
      mockPlatformKeys[idx] = { ...mockPlatformKeys[idx], ...body }
      return HttpResponse.json(mockPlatformKeys[idx])
    }
    return HttpResponse.json(null, { status: 404 })
  }),
  http.put(`${API_BASE_PATH}/keys/platform/:id/toggle`, async ({ params, request }) => {
    await delay(300)
    const body = (await request.json()) as { enabled: boolean }
    const idx = mockPlatformKeys.findIndex((k) => k.id === params.id)
    if (idx >= 0) {
      mockPlatformKeys[idx] = {
        ...mockPlatformKeys[idx],
        status: body.enabled ? 'active' : 'disabled',
      }
      return HttpResponse.json(mockPlatformKeys[idx])
    }
    return HttpResponse.json(null, { status: 404 })
  }),
  http.post(`${API_BASE_PATH}/keys/platform/:id/rotate`, async ({ params }) => {
    await delay(500)
    const idx = mockPlatformKeys.findIndex((k) => k.id === params.id)
    if (idx >= 0) {
      const fullKey = `tj-rot-${Date.now()}-demo-secret`
      mockPlatformKeys[idx] = {
        ...mockPlatformKeys[idx],
        keyPrefix: `${fullKey.slice(0, 12)}...`,
        fullKey,
      }
      return HttpResponse.json(mockPlatformKeys[idx])
    }
    return HttpResponse.json(null, { status: 404 })
  }),
  http.put(`${API_BASE_PATH}/keys/platform/:id/revoke`, async () => {
    await delay(300)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.delete(`${API_BASE_PATH}/keys/platform/:id`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/keys/approvals`, ({ request }) => {
    const url = new URL(request.url)
    const tab = url.searchParams.get('tab')
    const memberId = url.searchParams.get('memberId')
    let items = [...mockApprovals]
    if (tab === 'pending') {
      items = items.filter((a) => a.status === 'pending')
    } else if (tab === 'mine' && memberId) {
      items = items.filter((a) => a.applicantId === memberId)
    }
    return HttpResponse.json(items)
  }),
  http.post(`${API_BASE_PATH}/keys/approvals`, async ({ request }) => {
    await delay(400)
    const body = (await request.json()) as Record<string, unknown>
    const memberId = body.memberId as string
    const member = findMemberById(mockMembers, memberId)
    const approval = {
      id: `apv-${Date.now()}`,
      type: body.type as 'key' | 'quota',
      applicant: member?.name ?? '申请人',
      applicantId: memberId,
      department: member?.departmentName ?? '',
      reason: body.reason as string,
      requestedQuota: body.requestedQuota as number,
      requestedModels: body.requestedModels as string[],
      status: 'pending' as const,
      approver: null,
      createdAt: new Date().toISOString().slice(0, 16).replace('T', ' '),
      resolvedAt: null,
    }
    mockApprovals.push(approval)
    return HttpResponse.json(approval)
  }),
  http.get(`${API_BASE_PATH}/keys/approvals/:id/quota-check`, ({ params }) => {
    const approval = mockApprovals.find((a) => a.id === params.id)
    const requested = approval?.requestedQuota ?? 0
    const reservedPool = approval
      ? getReservedPoolForMember(mockBudgetTree, mockMembers, approval.applicantId)
      : 0
    return HttpResponse.json({
      sufficient: requested <= reservedPool,
      reservedPool,
      requested,
    })
  }),
  http.put(`${API_BASE_PATH}/keys/approvals/:id/approve`, async ({ params }) => {
    await delay(500)
    const idx = mockApprovals.findIndex((a) => a.id === params.id)
    if (idx < 0) {
      return HttpResponse.json(null, { status: 404 })
    }

    const approval = mockApprovals[idx]
    const reservedPool = getReservedPoolForMember(mockBudgetTree, mockMembers, approval.applicantId)
    if (approval.requestedQuota > reservedPool && approval.type === 'quota') {
      return HttpResponse.json({ message: 'Reserved pool insufficient' }, { status: 422 })
    }

    if (approval.type === 'key') {
      const member = mockMembers.find((m) => m.id === approval.applicantId)
      const fullKey = `tj-apv-${Date.now()}-demo-secret-key`
      mockPlatformKeys.push({
        id: `plk-apv-${Date.now()}`,
        name: `${approval.applicant}-审批 Key`,
        keyPrefix: `${fullKey.slice(0, 12)}...`,
        fullKey,
        memberId: approval.applicantId,
        memberName: member?.name ?? approval.applicant,
        appName: null,
        quota: approval.requestedQuota,
        used: 0,
        modelWhitelist: approval.requestedModels,
        status: 'active',
        createdAt: new Date().toISOString().slice(0, 10),
        expiresAt: null,
      })
    } else if (approval.type === 'quota') {
      const memberKey = mockPlatformKeys.find(
        (k) => k.memberId === approval.applicantId && k.status === 'active',
      )
      if (memberKey) {
        const keyIdx = mockPlatformKeys.findIndex((k) => k.id === memberKey.id)
        mockPlatformKeys[keyIdx] = {
          ...memberKey,
          quota: memberKey.quota + approval.requestedQuota,
        }
      }
    }

    mockApprovals[idx] = {
      ...approval,
      status: 'approved',
      approver: '李四',
      resolvedAt: new Date().toISOString().slice(0, 16).replace('T', ' '),
    }
    return HttpResponse.json(null, { status: 200 })
  }),
  http.put(`${API_BASE_PATH}/keys/approvals/:id/reject`, async ({ params, request }) => {
    await delay(500)
    const body = (await request.json()) as { reason?: string }
    const idx = mockApprovals.findIndex((a) => a.id === params.id)
    if (idx >= 0) {
      mockApprovals[idx] = {
        ...mockApprovals[idx],
        status: 'rejected',
        approver: '李四',
        rejectReason: body.reason ?? null,
        resolvedAt: new Date().toISOString().slice(0, 16).replace('T', ' '),
      }
    }
    return HttpResponse.json(null, { status: 200 })
  }),

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

  // ========== 审计日志 ==========
  http.get(`${API_BASE_PATH}/audit/operations`, ({ request }) => {
    const url = new URL(request.url)
    const action = url.searchParams.get('action')
    const page = parseIntParam(url.searchParams.get('page'), 1)
    const pageSize = parseIntParam(url.searchParams.get('pageSize'), 20)
    const items = action ? mockOperationLogs.filter((l) => l.action === action) : mockOperationLogs
    return HttpResponse.json(paginate(items, page, pageSize))
  }),
  http.get(`${API_BASE_PATH}/audit/calls`, ({ request }) => {
    const url = new URL(request.url)
    const model = url.searchParams.get('model')
    const status = url.searchParams.get('status')
    const page = parseIntParam(url.searchParams.get('page'), 1)
    const pageSize = parseIntParam(url.searchParams.get('pageSize'), 20)
    let items = mockCallLogs
    if (model) items = items.filter((l) => l.model === model)
    if (status) items = items.filter((l) => l.status === status)
    return HttpResponse.json(paginate(items, page, pageSize))
  }),
]

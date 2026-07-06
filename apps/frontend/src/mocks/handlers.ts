import { http, HttpResponse, delay } from 'msw'
import type { Department } from '../api/types'
import {
  mockDataSourceStatus,
  mockSyncConfig,
  mockSyncLogs,
  mockFieldMappings,
  mockDepartments,
  mockMembers,
  mockRoles,
  mockPermissions,
  mockBudgetTree,
  mockBudgetProjects,
  mockBudgetApprovals,
  mockAlertRules,
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

/** Collect a department ID and all descendant IDs */
function collectDepartmentIds(rootId: string, departments: Department[]): Set<string> {
  const ids = new Set<string>()
  function walk(nodes: Department[]) {
    for (const n of nodes) {
      if (n.id === rootId || ids.has(n.parentId ?? '')) {
        ids.add(n.id)
      }
      if (n.children) walk(n.children)
    }
  }
  ids.add(rootId)
  walk(departments)
  return ids
}

export const handlers = [
  // ========== 数据源 ==========
  http.get('/api/org/data-source/status', () => {
    return HttpResponse.json(mockDataSourceStatus)
  }),
  http.post('/api/org/data-source/test', async () => {
    await delay(1000)
    return HttpResponse.json({ success: true })
  }),
  http.put('/api/org/data-source', () => {
    mockDataSourceStatus.connected = true
    mockDataSourceStatus.platform = 'feishu'
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/org/data-source/search', ({ request }) => {
    const url = new URL(request.url)
    const keyword = url.searchParams.get('keyword') || ''
    if (!keyword) return HttpResponse.json({ name: '', department: '', mappingOk: false })
    return HttpResponse.json({
      name: keyword === '张三' ? '张三' : `${keyword}（模拟）`,
      department: '技术部 > 后端组',
      mappingOk: true,
    })
  }),
  http.post('/api/org/data-source/import', async () => {
    await delay(2000)
    return HttpResponse.json({
      successMembers: 120, successDepartments: 5,
      failures: [
        { id: 'f-1', name: '李四', employeeId: '10087', reason: '手机号为空' },
        { id: 'f-2', name: '王五', employeeId: '10088', reason: '部门不存在' },
      ],
    })
  }),
  http.post('/api/org/data-source/import/retry', async () => {
    await delay(500)
    return HttpResponse.json({ successMembers: 1, successDepartments: 0, failures: [] })
  }),
  http.get('/api/org/data-source/field-mappings', () => {
    return HttpResponse.json(mockFieldMappings)
  }),
  http.put('/api/org/data-source/field-mappings', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/org/data-source/field-mappings/test', async ({ request }) => {
    await delay(800)
    const url = new URL(request.url)
    const keyword = url.searchParams.get('keyword') || ''
    if (!keyword) {
      return HttpResponse.json({ success: false, preview: {}, errors: ['请输入搜索关键词'] })
    }
    return HttpResponse.json({
      success: true,
      preview: {
        姓名: keyword === '张三' ? '张三' : `${keyword}（模拟）`,
        手机号: '138****1234',
        邮箱: 'user@example.com',
        部门: '技术部 > 后端组',
        状态: '在职',
      },
      errors: [],
    })
  }),

  // ========== 同步 ==========
  http.get('/api/org/sync/config', () => {
    return HttpResponse.json(mockSyncConfig)
  }),
  http.put('/api/org/sync/config', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post('/api/org/sync/trigger', async () => {
    await delay(1500)
    return HttpResponse.json({ successMembers: 3, successDepartments: 0, failures: [] })
  }),
  http.get('/api/org/sync/logs', () => {
    return HttpResponse.json({ items: mockSyncLogs, total: 3, page: 1, pageSize: 10 })
  }),

  // ========== 部门 ==========
  http.get('/api/org/departments/tree', () => {
    return HttpResponse.json(mockDepartments)
  }),
  http.post('/api/org/departments', async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as { name: string; parentId: string }
    const newDept = { id: `dept-${Date.now()}`, name: body.name, parentId: body.parentId, memberCount: 0 }
    // Insert into tree
    const insertInto = (nodes: typeof mockDepartments): boolean => {
      for (const node of nodes) {
        if (node.id === body.parentId) {
          if (!node.children) node.children = []
          node.children.push(newDept)
          return true
        }
        if (node.children && insertInto(node.children)) return true
      }
      return false
    }
    insertInto(mockDepartments)
    return HttpResponse.json(newDept)
  }),
  http.put('/api/org/departments/:id', async ({ params, request }) => {
    await delay(200)
    const body = (await request.json()) as { name: string }
    const id = params.id as string
    // Update in tree
    const updateIn = (nodes: typeof mockDepartments): boolean => {
      for (const node of nodes) {
        if (node.id === id) { node.name = body.name; return true }
        if (node.children && updateIn(node.children)) return true
      }
      return false
    }
    updateIn(mockDepartments)
    // Also update departmentName in members
    for (const m of mockMembers) {
      if (m.departmentId === id) m.departmentName = body.name
    }
    return HttpResponse.json({ id, name: body.name, parentId: null, memberCount: 0 })
  }),
  http.delete('/api/org/departments/:id', ({ params }) => {
    const id = params.id as string
    const removeFrom = (nodes: typeof mockDepartments): boolean => {
      const idx = nodes.findIndex((n) => n.id === id)
      if (idx !== -1) { nodes.splice(idx, 1); return true }
      for (const node of nodes) {
        if (node.children && removeFrom(node.children)) return true
      }
      return false
    }
    removeFrom(mockDepartments)
    return HttpResponse.json(null, { status: 200 })
  }),

  // ========== 成员 ==========
  http.get('/api/org/members', ({ request }) => {
    const url = new URL(request.url)
    const deptId = url.searchParams.get('departmentId')
    const keyword = url.searchParams.get('keyword') || ''
    const page = Number(url.searchParams.get('page') || '1')
    const pageSize = Number(url.searchParams.get('pageSize') || '10')

    let items = [...mockMembers]

    // Filter by department — always show only direct members of this department
    if (deptId) {
      items = items.filter((m) => m.departmentId === deptId)
    }

    // Filter by keyword
    if (keyword) {
      const kw = keyword.toLowerCase()
      items = items.filter(
        (m) => m.name.toLowerCase().includes(kw) || m.phone.includes(kw) || m.email.toLowerCase().includes(kw)
      )
    }

    const total = items.length
    const start = (page - 1) * pageSize
    const paged = items.slice(start, start + pageSize)

    return HttpResponse.json({ items: paged, total, page, pageSize })
  }),
  http.post('/api/org/members', async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as Record<string, string>
    const newMember = {
      id: `m-${Date.now()}`,
      name: body.name,
      phone: body.phone,
      email: body.email,
      username: body.username || '',
      employeeId: body.employeeId || '',
      jobTitle: body.jobTitle || '',
      hireDate: body.hireDate || '',
      departmentId: body.departmentId,
      departmentName: body.departmentName,
      status: 'active' as const,
      roles: ['普通成员'],
      source: 'manual' as const,
    }
    mockMembers.push(newMember)
    // Update department member count
    updateDeptMemberCount(body.departmentId, 1, mockDepartments)
    return HttpResponse.json(newMember)
  }),
  http.put('/api/org/members/:id', async ({ params, request }) => {
    await delay(200)
    const id = params.id as string
    const body = (await request.json()) as Partial<typeof mockMembers[0]>
    const member = mockMembers.find((m) => m.id === id)
    if (member) {
      const oldDeptId = member.departmentId
      Object.assign(member, body)
      // If department changed, update counts
      if (body.departmentId && body.departmentId !== oldDeptId) {
        updateDeptMemberCount(oldDeptId, -1, mockDepartments)
        updateDeptMemberCount(body.departmentId, 1, mockDepartments)
      }
    }
    return HttpResponse.json(member)
  }),
  http.delete('/api/org/members', async ({ request }) => {
    await delay(200)
    const body = (await request.json()) as { ids: string[] }
    for (const id of body.ids) {
      const idx = mockMembers.findIndex((m) => m.id === id)
      if (idx !== -1) {
        updateDeptMemberCount(mockMembers[idx].departmentId, -1, mockDepartments)
        mockMembers.splice(idx, 1)
      }
    }
    return HttpResponse.json(null, { status: 200 })
  }),
  http.put('/api/org/members/status', async ({ request }) => {
    await delay(200)
    const body = (await request.json()) as { ids: string[]; status: string }
    for (const id of body.ids) {
      const member = mockMembers.find((m) => m.id === id)
      if (member) member.status = body.status as 'active' | 'inactive' | 'pending'
    }
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post('/api/org/members/transfer', async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as { ids: string[]; departmentId: string }
    const targetDeptName = findDeptName(body.departmentId, mockDepartments)
    for (const id of body.ids) {
      const member = mockMembers.find((m) => m.id === id)
      if (member) {
        updateDeptMemberCount(member.departmentId, -1, mockDepartments)
        member.departmentId = body.departmentId
        member.departmentName = targetDeptName
        updateDeptMemberCount(body.departmentId, 1, mockDepartments)
      }
    }
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post('/api/org/members/invite', async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as { email?: string; phone?: string }
    const newMember = {
      id: `m-${Date.now()}`,
      name: body.email || body.phone || '待激活用户',
      phone: body.phone || '',
      email: body.email || '',
      username: '',
      employeeId: '',
      jobTitle: '',
      hireDate: '',
      departmentId: 'dept-1',
      departmentName: '总公司',
      status: 'pending' as const,
      roles: ['普通成员'],
      source: 'invited' as const,
    }
    mockMembers.push(newMember)
    updateDeptMemberCount('dept-1', 1, mockDepartments)
    return HttpResponse.json(null, { status: 200 })
  }),

  // ========== 角色 ==========
  http.get('/api/org/roles', () => {
    return HttpResponse.json(mockRoles)
  }),
  http.post('/api/org/roles', async ({ request }) => {
    const body = (await request.json()) as { name: string; permissions: string[] }
    return HttpResponse.json({ id: `role-${Date.now()}`, ...body, type: 'custom', memberCount: 0 })
  }),
  http.put('/api/org/roles/:id', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(body)
  }),
  http.delete('/api/org/roles/:id', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/org/roles/:roleId/members', () => {
    return HttpResponse.json(mockMembers.slice(0, 2))
  }),
  http.post('/api/org/roles/:roleId/members', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.delete('/api/org/roles/:roleId/members/:memberId', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/org/permissions', () => {
    return HttpResponse.json(mockPermissions)
  }),

  http.get('/api/budget/tree', () => {
    return HttpResponse.json(mockBudgetTree)
  }),
  http.put('/api/budget/nodes/:id', async ({ request }) => {
    await delay(300)
    const body = await request.json()
    return HttpResponse.json({ id: 'dept-x', ...body as object })
  }),

  // Projects
  http.get('/api/budget/projects', () => {
    return HttpResponse.json(mockBudgetProjects)
  }),
  http.post('/api/budget/projects', async ({ request }) => {
    const data = await request.json() as Record<string, unknown>
    const project = { ...data, id: `proj-${Date.now()}`, consumed: 0 }
    mockBudgetProjects.push(project as typeof mockBudgetProjects[0])
    return HttpResponse.json(project, { status: 201 })
  }),
  http.put('/api/budget/projects/:id', async ({ params, request }) => {
    const data = await request.json() as Record<string, unknown>
    const idx = mockBudgetProjects.findIndex(p => p.id === params.id)
    if (idx >= 0) Object.assign(mockBudgetProjects[idx], data)
    return HttpResponse.json(mockBudgetProjects[idx])
  }),
  http.delete('/api/budget/projects/:id', ({ params }) => {
    const idx = mockBudgetProjects.findIndex(p => p.id === params.id)
    if (idx >= 0) mockBudgetProjects.splice(idx, 1)
    return HttpResponse.json(null, { status: 204 })
  }),

  // Approvals
  http.get('/api/budget/approvals', () => {
    return HttpResponse.json(mockBudgetApprovals)
  }),
  http.put('/api/budget/approvals/:id', async ({ params, request }) => {
    const data = await request.json() as Record<string, unknown>
    const idx = mockBudgetApprovals.findIndex(a => a.id === params.id)
    if (idx >= 0) Object.assign(mockBudgetApprovals[idx], data, { resolvedAt: new Date().toISOString() })
    return HttpResponse.json(mockBudgetApprovals[idx])
  }),

  http.get('/api/budget/alerts', () => {
    return HttpResponse.json(mockAlertRules)
  }),
  http.post('/api/budget/alerts', async ({ request }) => {
    await delay(300)
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ id: `alert-${Date.now()}`, ...body })
  }),
  http.put('/api/budget/alerts/:id', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(body)
  }),
  http.delete('/api/budget/alerts/:id', () => {
    return HttpResponse.json(null, { status: 200 })
  }),

  // ========== API-KEY 管理 ==========
  http.get('/api/keys/provider', () => {
    return HttpResponse.json(mockProviderKeys)
  }),
  http.post('/api/keys/provider', async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ id: `pk-${Date.now()}`, ...body, keyPrefix: 'sk-new...', status: 'active', balance: null, lastUsed: null, createdAt: '2026-06-19', rotateEnabled: false })
  }),
  http.put('/api/keys/provider/:id/toggle', async () => {
    await delay(300)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post('/api/keys/provider/:id/rotate', async () => {
    await delay(1000)
    return HttpResponse.json({ success: true })
  }),
  http.delete('/api/keys/provider/:id', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/keys/platform', ({ request }) => {
    const url = new URL(request.url)
    const departmentId = url.searchParams.get('departmentId')
    const type = url.searchParams.get('type')
    let items = mockPlatformKeys
    if (departmentId) {
      const childIds = collectDepartmentIds(departmentId, mockDepartments)
      items = items.filter(k => childIds.has(k.departmentId))
    }
    if (type) {
      items = items.filter(k => k.type === type)
    }
    return HttpResponse.json({ items, total: items.length, page: 1, pageSize: 20 })
  }),
  http.post('/api/keys/platform', async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ id: `plk-${Date.now()}`, ...body, keyPrefix: 'tj-new...', status: 'active', used: 0, createdAt: '2026-06-19', expiresAt: null })
  }),
  http.put('/api/keys/platform/:id/revoke', async () => {
    await delay(300)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.delete('/api/keys/platform/:id', () => {
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/keys/approvals', ({ request }) => {
    const url = new URL(request.url)
    const status = url.searchParams.get('status')
    const items = status ? mockApprovals.filter(a => a.status === status) : mockApprovals
    return HttpResponse.json(items)
  }),
  http.put('/api/keys/approvals/:id/approve', async () => {
    await delay(500)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.put('/api/keys/approvals/:id/reject', async () => {
    await delay(500)
    return HttpResponse.json(null, { status: 200 })
  }),

  // ========== 模型路由 ==========
  http.get('/api/models', () => {
    return HttpResponse.json(mockModels)
  }),
  http.post('/api/models', async ({ request }) => {
    const data = await request.json() as Record<string, unknown>
    const model = { ...data, id: `model-${Date.now()}` }
    mockModels.push(model as any)
    return HttpResponse.json(model, { status: 201 })
  }),
  http.put('/api/models/:id', async ({ params, request }) => {
    const data = await request.json() as Record<string, unknown>
    const idx = mockModels.findIndex(m => m.id === params.id)
    if (idx >= 0) Object.assign(mockModels[idx], data)
    return HttpResponse.json(mockModels[idx])
  }),
  http.delete('/api/models/:id', ({ params }) => {
    const idx = mockModels.findIndex(m => m.id === params.id)
    if (idx >= 0) mockModels.splice(idx, 1)
    return HttpResponse.json(null, { status: 204 })
  }),
  http.put('/api/models/:id/toggle', async () => {
    await delay(300)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get('/api/models/routing', () => {
    return HttpResponse.json(mockRoutingRules)
  }),
  http.put('/api/models/routing/:id', async ({ request }) => {
    await delay(300)
    const body = await request.json()
    return HttpResponse.json(body)
  }),

  // ========== 数据看板 ==========
  http.get('/api/dashboard/cost/summary', () => {
    return HttpResponse.json(mockCostSummary)
  }),
  http.get('/api/dashboard/cost/departments', () => {
    return HttpResponse.json(mockDepartmentCosts)
  }),
  http.get('/api/dashboard/cost/daily', () => {
    return HttpResponse.json(mockDailyCosts)
  }),
  http.get('/api/dashboard/cost/top', () => {
    return HttpResponse.json(mockTopConsumers)
  }),
  http.get('/api/dashboard/usage/models', () => {
    return HttpResponse.json(mockModelUsage)
  }),
  http.get('/api/dashboard/usage/teams', () => {
    return HttpResponse.json(mockTeamUsage)
  }),

  // ========== 审计日志 ==========
  http.get('/api/audit/operations', ({ request }) => {
    const url = new URL(request.url)
    const action = url.searchParams.get('action')
    const items = action ? mockOperationLogs.filter(l => l.action === action) : mockOperationLogs
    return HttpResponse.json({ items, total: items.length, page: 1, pageSize: 20 })
  }),
  http.get('/api/audit/calls', ({ request }) => {
    const url = new URL(request.url)
    const model = url.searchParams.get('model')
    const status = url.searchParams.get('status')
    let items = mockCallLogs
    if (model) items = items.filter(l => l.model === model)
    if (status) items = items.filter(l => l.status === status)
    return HttpResponse.json({ items, total: items.length, page: 1, pageSize: 20 })
  }),
]

// ========== Helper functions for stateful mock ==========

function updateDeptMemberCount(deptId: string, delta: number, tree: Department[]): boolean {
  for (const node of tree) {
    if (node.id === deptId) {
      node.memberCount = Math.max(0, node.memberCount + delta)
      return true
    }
    if (node.children && updateDeptMemberCount(deptId, delta, node.children)) return true
  }
  return false
}

function findDeptName(deptId: string, tree: Department[]): string {
  for (const node of tree) {
    if (node.id === deptId) return node.name
    if (node.children) {
      const name = findDeptName(deptId, node.children)
      if (name) return name
    }
  }
  return ''
}

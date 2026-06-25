import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import type { BatchImportRow } from '@/api/types'
import { flattenDepartmentTree, getDeptPath } from '@/lib/org'
import { paginate } from '../lib/paginate'
import { parseIntParam } from '../lib/parse'
import { filterMembersByDepartment } from '../lib/query'
import {
  mockDataSourceStatus,
  mockSyncConfig,
  mockSyncLogs,
  mockImportFailures,
  mockDepartments,
  mockMembers,
  mockRoles,
  mockPermissions,
  mockPlatformKeys,
} from '../data'
import { ROLE_MEMBER, countMembersByRole } from '../lib/member-factory'

function recalcRoleMemberCounts() {
  for (const role of mockRoles) {
    role.memberCount = countMembersByRole(mockMembers, role.name)
  }
}

export const orgHandlers = [
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
    const keyword = url.searchParams.get('keyword')?.trim() ?? ''
    if (!keyword) return HttpResponse.json({ name: '', department: '', mappingOk: false })
    const member = mockMembers.find(
      (m) => m.name.includes(keyword) || m.phone.includes(keyword) || m.email.includes(keyword),
    )
    if (!member) {
      return HttpResponse.json({ name: '', department: '', mappingOk: false })
    }
    const department = getDeptPath(mockDepartments, member.departmentId) ?? member.departmentName
    return HttpResponse.json({
      name: member.name,
      department,
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
    const role = {
      id: `role-${Date.now()}`,
      ...body,
      type: 'custom' as const,
      memberCount: 0,
    }
    mockRoles.push(role)
    return HttpResponse.json(role)
  }),
  http.put(`${API_BASE_PATH}/org/roles/:id`, async ({ params, request }) => {
    const body = (await request.json()) as { name: string; permissions: string[] }
    const idx = mockRoles.findIndex((r) => r.id === params.id)
    if (idx < 0) return HttpResponse.json({ message: 'Not found' }, { status: 404 })
    mockRoles[idx] = { ...mockRoles[idx], ...body }
    return HttpResponse.json(mockRoles[idx])
  }),
  http.delete(`${API_BASE_PATH}/org/roles/:id`, ({ params }) => {
    const idx = mockRoles.findIndex((r) => r.id === params.id)
    if (idx < 0) return HttpResponse.json(null, { status: 404 })
    const role = mockRoles[idx]
    if (role.type === 'preset') {
      return HttpResponse.json({ message: 'Cannot delete preset role' }, { status: 400 })
    }
    for (const member of mockMembers) {
      member.roles = member.roles.filter((name) => name !== role.name)
    }
    mockRoles.splice(idx, 1)
    recalcRoleMemberCounts()
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/org/roles/:roleId/members`, ({ params }) => {
    const role = mockRoles.find((r) => r.id === params.roleId)
    if (!role) return HttpResponse.json([])
    return HttpResponse.json(mockMembers.filter((m) => m.roles.includes(role.name)))
  }),
  http.post(`${API_BASE_PATH}/org/roles/:roleId/members`, async ({ params, request }) => {
    const body = (await request.json()) as { memberId: string }
    const role = mockRoles.find((r) => r.id === params.roleId)
    const member = mockMembers.find((m) => m.id === body.memberId)
    if (role && member && !member.roles.includes(role.name)) {
      member.roles = [...member.roles, role.name]
      recalcRoleMemberCounts()
    }
    return HttpResponse.json(null, { status: 200 })
  }),
  http.delete(`${API_BASE_PATH}/org/roles/:roleId/members/:memberId`, ({ params }) => {
    const role = mockRoles.find((r) => r.id === params.roleId)
    const member = mockMembers.find((m) => m.id === params.memberId)
    if (!role || !member) {
      return HttpResponse.json({ message: 'Not found' }, { status: 404 })
    }
    if (role.name === ROLE_MEMBER) {
      return HttpResponse.json({ message: 'Cannot remove base member role' }, { status: 400 })
    }
    member.roles = member.roles.filter((name) => name !== role.name)
    recalcRoleMemberCounts()
    return HttpResponse.json(null, { status: 200 })
  }),
  http.get(`${API_BASE_PATH}/org/permissions`, () => {
    return HttpResponse.json(mockPermissions)
  }),
]

import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { getReservedPoolForMember } from '../lib/budget-lookup'
import { findMemberById } from '../lib/query'
import {
  mockProviderKeys,
  mockPlatformKeys,
  mockApprovals,
  mockBudgetTree,
  mockMembers,
} from '../data'

export const keysHandlers = [
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
]

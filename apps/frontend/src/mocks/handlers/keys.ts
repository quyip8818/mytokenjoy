import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { MODEL_NOT_IN_DEPT_MESSAGE } from '@/lib/dashboard-constants'
import { getReservedPoolForMember } from '../lib/budget-lookup'
import {
  addPersonalQuota,
  buildQuotaSummary,
  getPersonalQuota,
  getQuotaRemaining,
} from '../lib/member-quota'
import { findMemberById } from '../lib/query'
import { resolveDeptAllowedModels } from '../lib/routing-helpers'
import {
  mockProviderKeys,
  mockPlatformKeys,
  mockApprovals,
  mockBudgetTree,
  mockMembers,
} from '../data'

function validateModelsForMember(memberId: string, models: string[]): string | null {
  const member = findMemberById(mockMembers, memberId)
  if (!member || models.length === 0) return null
  const allowed = resolveDeptAllowedModels(member.departmentId)
  const invalid = models.filter((m) => !allowed.includes(m))
  if (invalid.length > 0) return MODEL_NOT_IN_DEPT_MESSAGE
  return null
}

export const keysHandlers = [
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
    const reservedPool = getReservedPoolForMember(mockBudgetTree, mockMembers, memberId)
    return HttpResponse.json(buildQuotaSummary(memberId, reservedPool))
  }),
  http.post(`${API_BASE_PATH}/keys/platform`, async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as Record<string, unknown>
    const memberId = (body.memberId as string) ?? null
    if (!memberId) {
      return HttpResponse.json({ message: 'memberId required' }, { status: 400 })
    }
    const quota = body.quota as number
    const modelWhitelist = body.modelWhitelist as string[]
    const modelError = validateModelsForMember(memberId, modelWhitelist)
    if (modelError) {
      return HttpResponse.json({ message: modelError }, { status: 422 })
    }
    if (quota > getQuotaRemaining(memberId)) {
      return HttpResponse.json({ message: '额度不足，请先申请追加' }, { status: 422 })
    }
    const fullKey = `tj-${Date.now()}-demo-secret-key`
    const member = findMemberById(mockMembers, memberId)
    const newKey = {
      id: `plk-${Date.now()}`,
      name: body.name as string,
      keyPrefix: `${fullKey.slice(0, 12)}...`,
      fullKey,
      memberId,
      memberName: member?.name ?? null,
      appName: (body.appName as string) ?? null,
      quota,
      used: 0,
      modelWhitelist,
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
    if (idx < 0) return HttpResponse.json(null, { status: 404 })
    const existing = mockPlatformKeys[idx]
    const memberId = existing.memberId
    if (body.modelWhitelist && memberId) {
      const modelError = validateModelsForMember(memberId, body.modelWhitelist as string[])
      if (modelError) {
        return HttpResponse.json({ message: modelError }, { status: 422 })
      }
    }
    if (body.quota !== undefined && memberId) {
      const otherAllocated = mockPlatformKeys
        .filter((k) => k.memberId === memberId && k.status === 'active' && k.id !== existing.id)
        .reduce((sum, k) => sum + k.quota, 0)
      const newTotal = otherAllocated + (body.quota as number)
      if (newTotal > getPersonalQuota(memberId)) {
        return HttpResponse.json({ message: '额度不足，请先申请追加' }, { status: 422 })
      }
    }
    mockPlatformKeys[idx] = { ...existing, ...body }
    return HttpResponse.json(mockPlatformKeys[idx])
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
  http.delete(`${API_BASE_PATH}/keys/platform/:id`, ({ params }) => {
    const idx = mockPlatformKeys.findIndex((k) => k.id === params.id)
    if (idx >= 0) {
      mockPlatformKeys.splice(idx, 1)
    }
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
    const requestedModels = body.requestedModels as string[]
    const modelError = validateModelsForMember(memberId, requestedModels)
    if (modelError) {
      return HttpResponse.json({ message: modelError }, { status: 422 })
    }
    const member = findMemberById(mockMembers, memberId)
    const approval = {
      id: `apv-${Date.now()}`,
      type: body.type as 'key' | 'quota',
      applicant: member?.name ?? '申请人',
      applicantId: memberId,
      department: member?.departmentName ?? '',
      reason: body.reason as string,
      requestedQuota: body.requestedQuota as number,
      requestedModels,
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
      const keyQuota = approval.requestedQuota
      if (keyQuota > getQuotaRemaining(approval.applicantId)) {
        addPersonalQuota(approval.applicantId, keyQuota - getQuotaRemaining(approval.applicantId))
      }
      mockPlatformKeys.push({
        id: `plk-apv-${Date.now()}`,
        name: `${approval.applicant}-审批 Key`,
        keyPrefix: `${fullKey.slice(0, 12)}...`,
        fullKey,
        memberId: approval.applicantId,
        memberName: member?.name ?? approval.applicant,
        appName: null,
        quota: keyQuota,
        used: 0,
        modelWhitelist: approval.requestedModels,
        status: 'active',
        createdAt: new Date().toISOString().slice(0, 10),
        expiresAt: null,
      })
    } else if (approval.type === 'quota') {
      addPersonalQuota(approval.applicantId, approval.requestedQuota)
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

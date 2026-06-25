import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { getReservedPoolForMember } from '../../lib/budget-lookup'
import { addPersonalQuota, getQuotaRemaining } from '../../lib/member-quota'
import { findMemberById } from '../../lib/query'
import { mockApprovals, mockBudgetTree, mockMembers, mockPlatformKeys } from '../../data'
import { validateModelsForMember } from './validation'

export const approvalKeysHandlers = [
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
        budgetGroupId: null,
        budgetGroupName: null,
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

import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { getReservedPoolForMember } from '../../lib/budget-lookup'
import { validateGroupKeyQuota } from '../../lib/budget-group-quota'
import { buildQuotaSummary, getPersonalQuota, getQuotaRemaining } from '../../lib/member-quota'
import { findMemberById } from '../../lib/query'
import { mockBudgetGroups, mockBudgetTree, mockMembers, mockPlatformKeys } from '../../data'
import { validateModelsForMember } from './validation'

export const platformKeysHandlers = [
  http.get(`${API_BASE_PATH}/keys/platform`, ({ request }) => {
    const url = new URL(request.url)
    const memberId = url.searchParams.get('memberId')
    const budgetGroupId = url.searchParams.get('budgetGroupId')
    let items = [...mockPlatformKeys]
    if (memberId) {
      items = items.filter((k) => k.memberId === memberId)
    }
    if (budgetGroupId) {
      items = items.filter((k) => k.budgetGroupId === budgetGroupId)
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
    const budgetGroupId = (body.budgetGroupId as string) ?? null
    const quota = body.quota as number
    const modelWhitelist = body.modelWhitelist as string[]

    if (budgetGroupId) {
      const group = mockBudgetGroups.find((g) => g.id === budgetGroupId)
      if (!group) {
        return HttpResponse.json({ message: 'Budget group not found' }, { status: 404 })
      }
      const groupError = validateGroupKeyQuota(group, quota)
      if (groupError) {
        return HttpResponse.json({ message: groupError }, { status: 422 })
      }
      if (memberId) {
        const modelError = validateModelsForMember(memberId, modelWhitelist)
        if (modelError) {
          return HttpResponse.json({ message: modelError }, { status: 422 })
        }
      }
    } else {
      if (!memberId) {
        return HttpResponse.json({ message: 'memberId required' }, { status: 400 })
      }
      const modelError = validateModelsForMember(memberId, modelWhitelist)
      if (modelError) {
        return HttpResponse.json({ message: modelError }, { status: 422 })
      }
      if (quota > getQuotaRemaining(memberId)) {
        return HttpResponse.json({ message: '额度不足，请先申请追加' }, { status: 422 })
      }
    }

    const fullKey = `tj-${Date.now()}-demo-secret-key`
    const member = memberId ? findMemberById(mockMembers, memberId) : undefined
    const group = budgetGroupId ? mockBudgetGroups.find((g) => g.id === budgetGroupId) : undefined
    const newKey = {
      id: `plk-${Date.now()}`,
      name: body.name as string,
      keyPrefix: `${fullKey.slice(0, 12)}...`,
      fullKey,
      memberId,
      memberName: member?.name ?? null,
      appName: (body.appName as string) ?? null,
      budgetGroupId,
      budgetGroupName: group?.name ?? null,
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
    const budgetGroupId = existing.budgetGroupId
    if (body.modelWhitelist && memberId) {
      const modelError = validateModelsForMember(memberId, body.modelWhitelist as string[])
      if (modelError) {
        return HttpResponse.json({ message: modelError }, { status: 422 })
      }
    }
    if (body.quota !== undefined) {
      if (budgetGroupId) {
        const group = mockBudgetGroups.find((g) => g.id === budgetGroupId)
        if (!group) {
          return HttpResponse.json({ message: 'Budget group not found' }, { status: 404 })
        }
        const groupError = validateGroupKeyQuota(group, body.quota as number, existing.id)
        if (groupError) {
          return HttpResponse.json({ message: groupError }, { status: 422 })
        }
      } else if (memberId) {
        const otherAllocated = mockPlatformKeys
          .filter(
            (k) =>
              k.memberId === memberId &&
              !k.budgetGroupId &&
              k.status === 'active' &&
              k.id !== existing.id,
          )
          .reduce((sum, k) => sum + k.quota, 0)
        const newTotal = otherAllocated + (body.quota as number)
        if (newTotal > getPersonalQuota(memberId)) {
          return HttpResponse.json({ message: '额度不足，请先申请追加' }, { status: 422 })
        }
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
]

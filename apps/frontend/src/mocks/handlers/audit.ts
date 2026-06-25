import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import type { AuditSettings } from '@/api/types'
import { paginate } from '../lib/paginate'
import { parseIntParam } from '../lib/parse'
import { filterByDateRange, filterByKeyword } from '../lib/audit-filter'
import { mockOperationLogs, mockCallLogs } from '../data'

let mockAuditSettings: AuditSettings = {
  contentRetentionEnabled: true,
}

export const auditHandlers = [
  http.get(`${API_BASE_PATH}/audit/settings`, () => {
    return HttpResponse.json(mockAuditSettings)
  }),
  http.put(`${API_BASE_PATH}/audit/settings`, async ({ request }) => {
    const body = (await request.json()) as AuditSettings
    mockAuditSettings = { ...mockAuditSettings, ...body }
    return HttpResponse.json(mockAuditSettings)
  }),
  http.get(`${API_BASE_PATH}/audit/operations`, ({ request }) => {
    const url = new URL(request.url)
    const action = url.searchParams.get('action')
    const operatorId = url.searchParams.get('operatorId')
    const keyword = url.searchParams.get('keyword') ?? undefined
    const from = url.searchParams.get('from') ?? undefined
    const to = url.searchParams.get('to') ?? undefined
    const page = parseIntParam(url.searchParams.get('page'), 1)
    const pageSize = parseIntParam(url.searchParams.get('pageSize'), 20)
    let items = [...mockOperationLogs]
    if (action) items = items.filter((l) => l.action === action)
    if (operatorId) items = items.filter((l) => l.operatorId === operatorId)
    items = filterByDateRange(items, from, to)
    items = filterByKeyword(items, keyword, ['detail', 'target', 'operator'])
    return HttpResponse.json(paginate(items, page, pageSize))
  }),
  http.get(`${API_BASE_PATH}/audit/calls`, ({ request }) => {
    const url = new URL(request.url)
    const model = url.searchParams.get('model')
    const status = url.searchParams.get('status')
    const callerId = url.searchParams.get('callerId')
    const keyword = url.searchParams.get('keyword') ?? undefined
    const from = url.searchParams.get('from') ?? undefined
    const to = url.searchParams.get('to') ?? undefined
    const page = parseIntParam(url.searchParams.get('page'), 1)
    const pageSize = parseIntParam(url.searchParams.get('pageSize'), 20)
    let items = [...mockCallLogs]
    if (model) items = items.filter((l) => l.model === model)
    if (status) items = items.filter((l) => l.status === status)
    if (callerId) items = items.filter((l) => l.callerId === callerId)
    items = filterByDateRange(items, from, to)
    items = filterByKeyword(items, keyword, ['inputPreview', 'outputPreview', 'caller', 'model'])
    return HttpResponse.json(paginate(items, page, pageSize))
  }),
]

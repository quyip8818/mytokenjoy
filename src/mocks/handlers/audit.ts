import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { paginate } from '../lib/paginate'
import { parseIntParam } from '../lib/parse'
import { mockOperationLogs, mockCallLogs } from '../data'

export const auditHandlers = [
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

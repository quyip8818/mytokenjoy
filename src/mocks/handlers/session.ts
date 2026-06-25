import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { resolveMemberPermissions, isReadOnlySession } from '@/lib/permissions'
import { mockMembers, mockRoles } from '../data'
import { findMemberById } from '../lib/query'

export const sessionHandlers = [
  http.get(`${API_BASE_PATH}/session`, ({ request }) => {
    const url = new URL(request.url)
    const memberId = url.searchParams.get('memberId')
    if (!memberId) {
      return HttpResponse.json({ message: 'memberId is required' }, { status: 400 })
    }

    const member = findMemberById(mockMembers, memberId)
    if (!member) {
      return HttpResponse.json({ message: 'Member not found' }, { status: 404 })
    }

    const permissions = resolveMemberPermissions(member, mockRoles)
    return HttpResponse.json({
      member,
      permissions,
      readOnly: isReadOnlySession(permissions),
    })
  }),
]

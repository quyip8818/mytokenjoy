import { http, HttpResponse } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { SessionContextSchema } from '@/api/schemas/session'
import { resolveMemberPermissions, isReadOnlySession } from '@/lib/permissions'
import { DEFAULT_DEMO_MEMBER_ID } from '@/features/demo/roles/constants'
import { mockMembers, mockRoles } from '../data'
import { findMemberById } from '../lib/query'

const SESSION_COOKIE = 'tokenjoy_session_member'

function buildSessionResponse(memberId: string) {
  const member = findMemberById(mockMembers, memberId)
  if (!member) {
    return null
  }
  const permissions = resolveMemberPermissions(member, mockRoles)
  return {
    member,
    permissions,
    readOnly: isReadOnlySession(permissions),
  }
}

function jsonSession(session: NonNullable<ReturnType<typeof buildSessionResponse>>) {
  const parsed = SessionContextSchema.safeParse(session)
  if (!parsed.success) {
    return HttpResponse.json({ message: 'Invalid session payload' }, { status: 500 })
  }
  return HttpResponse.json(parsed.data)
}

function resolveMemberIdFromRequest(request: Request): string | null {
  const url = new URL(request.url)
  const queryMemberId = url.searchParams.get('memberId')
  if (queryMemberId) {
    return queryMemberId
  }

  const cookieHeader = request.headers.get('cookie') ?? ''
  const cookieMatch = cookieHeader.match(new RegExp(`${SESSION_COOKIE}=([^;]+)`))
  if (cookieMatch?.[1]) {
    return decodeURIComponent(cookieMatch[1])
  }

  const authorization = request.headers.get('authorization')
  if (authorization?.startsWith('Bearer ')) {
    const token = authorization.slice('Bearer '.length).trim()
    if (token) {
      return token
    }
  }

  return DEFAULT_DEMO_MEMBER_ID
}

export const sessionHandlers = [
  http.get(`${API_BASE_PATH}/session`, ({ request }) => {
    const url = new URL(request.url)
    const hasMemberIdQuery = url.searchParams.has('memberId')

    if (hasMemberIdQuery) {
      const memberId = url.searchParams.get('memberId')
      if (!memberId) {
        return HttpResponse.json({ message: 'memberId is required' }, { status: 400 })
      }
      const session = buildSessionResponse(memberId)
      if (!session) {
        return HttpResponse.json({ message: 'Member not found' }, { status: 404 })
      }
      return jsonSession(session)
    }

    const memberId = resolveMemberIdFromRequest(request)
    if (!memberId) {
      return HttpResponse.json({ message: 'Unauthorized' }, { status: 401 })
    }

    const session = buildSessionResponse(memberId)
    if (!session) {
      return HttpResponse.json({ message: 'Member not found' }, { status: 404 })
    }

    return jsonSession(session)
  }),
]

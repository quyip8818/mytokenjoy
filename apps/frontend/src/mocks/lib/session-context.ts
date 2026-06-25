import type { Member } from '@/api/types'
import { findMemberById } from './query'

const DEMO_MEMBER_ID_HEADER = 'X-Demo-Member-Id'

export function resolveDemoMemberName(request: Request, members: Member[]): string {
  const memberId = request.headers.get(DEMO_MEMBER_ID_HEADER)
  if (!memberId) return '审批人'
  return findMemberById(members, memberId)?.name ?? '审批人'
}

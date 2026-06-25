import { MODEL_NOT_IN_DEPT_MESSAGE } from '@/lib/dashboard-constants'
import { findMemberById } from '../../lib/query'
import { resolveDeptAllowedModels } from '../../lib/routing-helpers'
import { mockMembers } from '../../data'

export function validateModelsForMember(memberId: string, models: string[]): string | null {
  const member = findMemberById(mockMembers, memberId)
  if (!member || models.length === 0) return null
  const allowed = resolveDeptAllowedModels(member.departmentId)
  const invalid = models.filter((m) => !allowed.includes(m))
  if (invalid.length > 0) return MODEL_NOT_IN_DEPT_MESSAGE
  return null
}

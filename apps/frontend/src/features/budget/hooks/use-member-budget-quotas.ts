import { useCallback } from 'react'
import type { MemberBudgetQuota } from '@/api/types'
import { useAsyncFetch } from './use-async-fetch'

const emptyMemberBudgets: MemberBudgetQuota[] = []

export function useMemberBudgetQuotas(
  departmentId: string,
  getMemberBudgets: (departmentId: string) => Promise<MemberBudgetQuota[]>,
  enabled = true,
) {
  const fetchMembers = useCallback(
    () => getMemberBudgets(departmentId).then((data) => data ?? emptyMemberBudgets),
    [departmentId, getMemberBudgets],
  )

  return useAsyncFetch(enabled ? departmentId : '', fetchMembers, enabled, emptyMemberBudgets)
}

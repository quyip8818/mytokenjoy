import { useCallback } from 'react'
import type { MemberBudget } from '@/api/types'
import { useAsyncFetch } from './use-async-fetch'

const emptyMemberBudgets: MemberBudget[] = []

export function useMemberBudgets(
  departmentId: string,
  getMemberBudgets: (departmentId: string) => Promise<MemberBudget[]>,
  enabled = true,
) {
  const fetchMembers = useCallback(
    () => getMemberBudgets(departmentId).then((data) => data ?? emptyMemberBudgets),
    [departmentId, getMemberBudgets],
  )

  return useAsyncFetch(enabled ? departmentId : '', fetchMembers, enabled, emptyMemberBudgets)
}

import { useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { Member } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'

type UseBudgetDepartmentMembersOptions = {
  injectedApis?: AppApis
  departmentId?: string
}

export function useBudgetDepartmentMembers({
  injectedApis,
  departmentId,
}: UseBudgetDepartmentMembersOptions) {
  const departmentMembersQuery = useMemo(
    () => ({
      departmentId,
      page: 1,
      pageSize: 200,
    }),
    [departmentId],
  )

  const { data: departmentMembersResult, loading: departmentMembersLoading } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.members(departmentMembersQuery),
    queryFn: (api) => api.memberApi.list(departmentMembersQuery),
    enabled: Boolean(departmentId),
  })

  const departmentMembers = useMemo(
    () => departmentMembersResult?.items ?? [],
    [departmentMembersResult?.items],
  )

  return { departmentMembers, departmentMembersLoading }
}

export function filterProjectMembers(departmentMembers: Member[], memberIds: string[]): Member[] {
  return departmentMembers.filter((member) => memberIds.includes(member.id))
}

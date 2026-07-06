import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'

export function useAuditMemberOptions(injectedApis?: AppApis) {
  const { data: members = [], loading } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.audit.members(),
    queryFn: (apis) => apis.memberApi.list({ page: 1, pageSize: 100 }).then((res) => res.items),
  })

  return { members, loading }
}

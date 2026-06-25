import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useAuditMemberOptions(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis

  const { data: members = [], loading } = useAsyncResource(
    () => apis.memberApi.list({ page: 1, pageSize: 100 }).then((res) => res.items),
    [apis],
  )

  return { members, loading }
}

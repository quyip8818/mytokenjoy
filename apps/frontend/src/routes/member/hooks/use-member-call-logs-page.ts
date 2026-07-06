import { useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import { useSession } from '@/features/session'

export function useMemberCallLogsPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { memberId } = useSession()
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.audit.calls({ callerId: memberId, page, pageSize }),
    queryFn: (api) =>
      api.auditApi.getCalls({
        callerId: memberId,
        page,
        pageSize,
      }),
    enabled: Boolean(memberId),
  })

  const logs = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return {
    logs,
    total,
    page,
    pageSize,
    totalPages,
    loading,
    error,
    refresh,
    setPage,
  }
}

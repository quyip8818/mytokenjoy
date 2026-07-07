import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useSession } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'

export const APPROVAL_PENDING_POLL_INTERVAL_MS = 30_000

export interface UseApprovalPendingCountQueryOptions {
  injectedApis?: AppApis
  poll?: boolean
}

export function useApprovalPendingCountQuery(options: UseApprovalPendingCountQueryOptions = {}) {
  const apis = useInjectedApis(options.injectedApis)
  const { permissions } = useSession()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)

  return useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.org.approvalPendingCount(),
    queryFn: async (a) => {
      if (!canApprove) return 0
      const items = await a.approvalApi.list({ tab: 'pending' })
      return items.length
    },
    enabled: canApprove,
    refetchInterval: options.poll && canApprove ? APPROVAL_PENDING_POLL_INTERVAL_MS : false,
  })
}

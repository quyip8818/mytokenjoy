import { useApis } from '@/api/use-apis'
import { useSession } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import { queryKeys, useInjectedQuery } from '@/features/query'

const APPROVAL_POLL_INTERVAL_MS = 30_000

export function useDemoApprovalPendingCount(): number {
  const apis = useApis()
  const { permissions } = useSession()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)

  const { data: count = 0 } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.org.approvalPendingCount(),
    queryFn: async (a) => {
      if (!canApprove) return 0
      const items = await a.approvalApi.list({ tab: 'pending' })
      return items.filter((item) => item.status === 'pending').length
    },
    enabled: canApprove,
    refetchInterval: canApprove ? APPROVAL_POLL_INTERVAL_MS : false,
  })

  return count
}

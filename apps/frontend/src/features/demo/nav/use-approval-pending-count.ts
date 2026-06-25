import { useApis } from '@/api/use-apis'
import { useSession } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useDemoApprovalPendingCount(): number {
  const apis = useApis()
  const { permissions } = useSession()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)

  const { data: count = 0 } = useAsyncResource(async () => {
    if (!canApprove) return 0
    const items = await apis.approvalApi.list({ tab: 'pending' })
    return items.filter((a) => a.status === 'pending').length
  }, [apis, canApprove])

  return count
}

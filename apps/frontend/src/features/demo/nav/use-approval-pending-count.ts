import { useDemoRole } from '@/features/demo/roles/use-demo-role'
import { PERMISSION } from '@/lib/permissions'
import { approvalApi } from '@/api/keys'
import { useAsyncResource } from '@/hooks/use-async-resource'

export function useApprovalPendingCount(): number {
  const { permissions } = useDemoRole()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)

  const { data: count = 0 } = useAsyncResource(async () => {
    if (!canApprove) return 0
    const items = await approvalApi.list({ tab: 'pending' })
    return items.filter((a) => a.status === 'pending').length
  }, [canApprove])

  return count
}

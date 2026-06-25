import { useApprovalPendingCountQuery } from '@/hooks/use-approval-pending-count-query'

export function useDemoApprovalPendingCount(): number {
  const { data: count = 0 } = useApprovalPendingCountQuery({ poll: true })
  return count
}

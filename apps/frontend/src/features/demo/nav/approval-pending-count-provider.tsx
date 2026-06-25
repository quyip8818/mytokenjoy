import type { ReactNode } from 'react'
import { ApprovalPendingCountContext } from '@/hooks/use-approval-pending-count'
import { useDemoApprovalPendingCount } from './use-approval-pending-count'

export function DemoApprovalPendingCountProvider({ children }: { children: ReactNode }) {
  const count = useDemoApprovalPendingCount()

  return (
    <ApprovalPendingCountContext.Provider value={count}>
      {children}
    </ApprovalPendingCountContext.Provider>
  )
}

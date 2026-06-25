import { createContext, useContext } from 'react'

export const ApprovalPendingCountContext = createContext<number>(0)

export function useApprovalPendingCount(): number {
  return useContext(ApprovalPendingCountContext)
}

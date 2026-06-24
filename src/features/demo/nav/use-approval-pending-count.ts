import { useEffect, useState } from 'react'
import { approvalApi } from '@/api/keys'
import { useDemoRole } from '@/features/demo/roles/use-demo-role'

export function useApprovalPendingCount(): number {
  const { role } = useDemoRole()
  const [count, setCount] = useState(0)
  const shouldFetch = role === 'admin' || role === 'tl'

  useEffect(() => {
    if (!shouldFetch) return
    let cancelled = false
    void approvalApi.list({ tab: 'pending' }).then((items) => {
      if (!cancelled) setCount(items.filter((a) => a.status === 'pending').length)
    })
    return () => {
      cancelled = true
    }
  }, [shouldFetch, role])

  return shouldFetch ? count : 0
}

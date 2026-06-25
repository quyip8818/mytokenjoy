import type { ReactNode } from 'react'
import { ErrorState } from '@/components/ui/error-state'
import { useDemoRole } from './roles/use-demo-role'

export function DemoSessionGate({ children }: { children: ReactNode }) {
  const { sessionError, loading, refreshSession } = useDemoRole()

  if (loading) {
    return null
  }

  if (sessionError) {
    return (
      <div className="flex min-h-screen items-center justify-center p-8">
        <ErrorState message={sessionError.message} onRetry={() => void refreshSession()} />
      </div>
    )
  }

  return children
}

import type { ReactNode } from 'react'
import { ErrorState } from '@/components/ui/error-state'
import { useSession } from './use-session'

export function SessionGate({ children }: { children: ReactNode }) {
  const { sessionError, loading, refreshSession } = useSession()

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

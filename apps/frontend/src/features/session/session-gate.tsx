import { useEffect, useRef, type ReactNode } from 'react'
import { ApiError } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'
import { ErrorState } from '@/components/ui/error-state'
import { useSession } from './use-session'

export function SessionGate({ children }: { children: ReactNode }) {
  const { sessionError, loading, refreshSession } = useSession()
  const hasRedirected = useRef(false)

  const isUnauthorized = sessionError instanceof ApiError && sessionError.status === 401

  useEffect(() => {
    if (isUnauthorized && !hasRedirected.current) {
      hasRedirected.current = true
      window.location.replace(LOGIN_PATH)
    }
  }, [isUnauthorized])

  if (loading) {
    return null
  }

  if (isUnauthorized) {
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

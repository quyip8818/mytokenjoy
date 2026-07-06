import { useEffect, type ReactNode } from 'react'
import { useNavigate } from 'react-router'
import { ApiError } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'
import { ErrorState } from '@/components/ui/error-state'
import { useSession } from './use-session'

export function SessionGate({ children }: { children: ReactNode }) {
  const { sessionError, loading, refreshSession } = useSession()
  const navigate = useNavigate()

  useEffect(() => {
    if (sessionError instanceof ApiError && sessionError.status === 401) {
      navigate(LOGIN_PATH, { replace: true })
    }
  }, [sessionError, navigate])

  if (loading) {
    return null
  }

  if (sessionError instanceof ApiError && sessionError.status === 401) {
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

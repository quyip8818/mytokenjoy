import { Navigate } from 'react-router'
import { useSession } from './use-session'

export function SessionGate({ children }: { children: React.ReactNode }) {
  const user = useSession((s) => s.user)
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

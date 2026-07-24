import { useLocation } from 'react-router'
import { canAccessRoute } from '@/lib/permissions'
import { useSession } from './use-session'

export function useRouteAccess() {
  const location = useLocation()
  const { permissions, loading } = useSession()

  const canAccess = canAccessRoute(location.pathname, permissions)

  return {
    pathname: location.pathname,
    permissions,
    loading,
    canAccess,
  }
}

import { useLocation } from 'react-router'
import { useSession } from '@/features/session'
import { canAccessCurrentRoute } from '@/lib/route-access'

export function useRouteAccess() {
  const location = useLocation()
  const { permissions, loading } = useSession()

  const canAccess = canAccessCurrentRoute(location.pathname, permissions)

  return {
    pathname: location.pathname,
    permissions,
    loading,
    canAccess,
  }
}

/* eslint-disable react-refresh/only-export-components */
import { createRoute, Navigate } from '@tanstack/react-router'
import { useSession } from '@/features/session'
import { getDefaultHomePath } from '@/lib/permissions'
import { EmptyState } from '@/components/ui/empty-state'
import { authLayoutRoute } from '../auth-layout'

function HomeRedirect() {
  const { permissions, loading } = useSession()

  if (loading) return null

  if (permissions.length === 0) {
    return (
      <EmptyState
        title="No accessible pages"
        description="Your account has no permissions assigned. Contact an administrator."
        className="m-6"
      />
    )
  }

  const homePath = getDefaultHomePath(permissions)
  if (!homePath) {
    return (
      <EmptyState
        title="No default page available"
        description="Your permissions do not match any configured home page. Use the sidebar to navigate."
        className="m-6"
      />
    )
  }

  return <Navigate to={homePath} replace />
}

export const homeRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/',
  component: HomeRedirect,
})

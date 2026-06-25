import { Navigate } from 'react-router'
import { EmptyState } from '@/components/ui/empty-state'
import { getDefaultHomePath } from '@/lib/permissions'
import { useDemoRole } from './use-demo-role'

export function HomeRedirect() {
  const { permissions, loading } = useDemoRole()

  if (loading) {
    return null
  }

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

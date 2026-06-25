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

  return <Navigate to={getDefaultHomePath(permissions)} replace />
}

import { useSession } from '@/features/session'
import { hasPermission, canWriteSession, type PermissionKey } from '@/lib/permissions'

export function usePermissions() {
  const { permissions, readOnly, loading } = useSession()

  return {
    permissions,
    readOnly,
    loading,
    has: (required: PermissionKey | PermissionKey[]) => hasPermission(permissions, required),
    canWrite: canWriteSession(permissions, readOnly),
  }
}

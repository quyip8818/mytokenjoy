import { useDemoRole } from '@/features/demo/roles/use-demo-role'
import { hasPermission, canWriteSession, type PermissionKey } from '@/lib/permissions'

export function usePermissions() {
  const { permissions, readOnly, loading } = useDemoRole()

  return {
    permissions,
    readOnly,
    loading,
    has: (required: PermissionKey | PermissionKey[]) => hasPermission(permissions, required),
    canWrite: canWriteSession(permissions),
  }
}

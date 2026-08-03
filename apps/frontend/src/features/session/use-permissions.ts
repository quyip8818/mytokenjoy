import { useMemo } from 'react'
import { useSession } from './use-session'
import {
  hasPermission,
  canWriteSession,
  expandHierarchy,
  type PermissionKey,
} from '@/lib/permissions'

export function usePermissions() {
  const { permissions: raw, readOnly, loading } = useSession()
  const permissions = useMemo(() => expandHierarchy(raw), [raw])

  return {
    permissions,
    readOnly,
    loading,
    has: (required: PermissionKey | PermissionKey[]) => hasPermission(permissions, required),
    canWrite: canWriteSession(permissions, readOnly),
  }
}

import type { ReactNode } from 'react'
import { usePermissions } from '../use-permissions'
import type { PermissionKey } from '@/lib/permissions'

interface PermissionGateProps {
  permission?: PermissionKey | PermissionKey[]
  write?: boolean
  children: ReactNode
  fallback?: ReactNode
}

export function PermissionGate({
  permission,
  write = false,
  children,
  fallback = null,
}: PermissionGateProps) {
  const { has, canWrite, loading } = usePermissions()

  if (loading) return null
  if (write && !canWrite) return <>{fallback}</>
  if (permission && !has(permission)) return <>{fallback}</>

  return <>{children}</>
}

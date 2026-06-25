import { canAccessRoute as canAccessRouteByPermissions } from '@/lib/permissions'

export function canAccessCurrentRoute(pathname: string, permissions: readonly string[]): boolean {
  return canAccessRouteByPermissions(pathname, permissions)
}

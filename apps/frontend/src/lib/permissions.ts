import { PERMISSION, type PermissionKey } from '@/lib/permission-keys'
import { HOME_PATH_CANDIDATES, ROUTE_META, type RoutePath, routePermissions } from '@/config/routes'

export { PERMISSION, type PermissionKey } from '@/lib/permission-keys'

export const ALL_PERMISSIONS: PermissionKey[] = Object.values(PERMISSION)

/**
 * Returns true if the user holds ANY of the required permissions (OR semantics).
 * Pass a single key or array of keys.
 */
export function hasPermission(
  permissions: readonly string[],
  required: PermissionKey | PermissionKey[],
): boolean {
  const requiredList = Array.isArray(required) ? required : [required]
  return requiredList.some((p) => permissions.includes(p))
}

export function isReadOnlySession(permissions: readonly string[], readOnly: boolean): boolean {
  if (permissions.includes('*' as PermissionKey)) return false
  return readOnly
}

export function canWriteSession(permissions: readonly string[], readOnly: boolean): boolean {
  return !isReadOnlySession(permissions, readOnly)
}

export function getDefaultHomePath(permissions: readonly string[]): RoutePath | null {
  for (const path of HOME_PATH_CANDIDATES) {
    if (hasPermission(permissions, routePermissions(path))) return path
  }
  return null
}

export function getRouteRequiredPermissions(pathname: string): PermissionKey[] | null {
  const match = ROUTE_META.find((route) => pathname.startsWith(route.path))
  return match ? [...match.requiredPermissions] : null
}

export function canAccessRoute(pathname: string, permissions: readonly string[]): boolean {
  const required = getRouteRequiredPermissions(pathname)
  if (!required) return true
  return hasPermission(permissions, required)
}

import { PERMISSION, type PermissionKey } from '@/lib/permission-keys'
import { HOME_PATH_CANDIDATES, ROUTE_META, type RoutePath, routePermissions } from '@/config/routes'

export { PERMISSION, type PermissionKey } from '@/lib/permission-keys'

export const ALL_PERMISSIONS: PermissionKey[] = Object.values(PERMISSION)

/**
 * Permission hierarchy: higher-level permissions imply lower-level ones.
 * ponytail: static map mirrors manifest.json hierarchy. Keep in sync when adding new domains.
 * Upgrade path: generate from manifest.json at build time.
 */
const HIERARCHY: Record<string, string[]> = {
  'org:admin': ['org:manage', 'org:read'],
  'org:manage': ['org:read'],
  'budget:admin': ['budget:manage', 'budget:read'],
  'budget:manage': ['budget:read'],
  'model:manage': ['model:read'],
  'keys:admin': ['keys:manage', 'keys:read'],
  'keys:manage': ['keys:read'],
  'billing:manage': ['billing:read'],
  'platform:admin': ['platform:manage', 'platform:read'],
  'platform:manage': ['platform:read'],
}

/**
 * Expands permissions by hierarchy. E.g. ['platform:admin'] → includes platform:manage, platform:read.
 * Defense-in-depth: backend already expands, but frontend does it too for safety.
 */
export function expandHierarchy(perms: readonly string[]): string[] {
  const result = new Set(perms)
  let changed = true
  while (changed) {
    changed = false
    for (const p of result) {
      const implied = HIERARCHY[p]
      if (!implied) continue
      for (const ip of implied) {
        if (!result.has(ip)) {
          result.add(ip)
          changed = true
        }
      }
    }
  }
  return [...result]
}

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
  if (permissions.includes('*')) return false
  return readOnly
}

export function canWriteSession(permissions: readonly string[], readOnly: boolean): boolean {
  return !isReadOnlySession(permissions, readOnly)
}

export function getDefaultHomePath(permissions: readonly string[]): RoutePath | null {
  for (const path of HOME_PATH_CANDIDATES) {
    const required = routePermissions(path)
    if (required.length === 0 || hasPermission(permissions, required)) return path
  }
  return null
}

export function getRouteRequiredPermissions(pathname: string): PermissionKey[] | null {
  const match = ROUTE_META.find((route) => pathname.startsWith(route.path))
  return match ? [...match.requiredPermissions] : null
}

export function canAccessRoute(pathname: string, permissions: readonly string[]): boolean {
  const required = getRouteRequiredPermissions(pathname)
  if (!required || required.length === 0) return true
  return hasPermission(permissions, required)
}

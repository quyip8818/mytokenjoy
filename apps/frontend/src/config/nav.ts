import type { LucideIcon } from 'lucide-react'
import { type PermissionKey, hasPermission } from '@/lib/permissions'
import { NAV_GROUP_LAYOUT, ROUTE_META, getRouteMeta, type RoutePath } from '@/config/routes'

export interface NavItem {
  label: string
  path: string
  icon: LucideIcon
  requiredPermissions: PermissionKey[]
  badgeKey?: 'approvalPending'
}

export interface NavGroup {
  group: string
  items: NavItem[]
  collapsed?: boolean
}

function toNavItem(path: RoutePath): NavItem {
  const meta = getRouteMeta(path)
  return {
    path: meta.path,
    label: meta.label,
    icon: meta.icon,
    requiredPermissions: [...meta.requiredPermissions],
    badgeKey: meta.badgeKey,
  }
}

export const NAV_GROUPS: NavGroup[] = NAV_GROUP_LAYOUT.map((layout) => ({
  group: layout.group,
  collapsed: layout.collapsed,
  items: layout.paths.map(toNavItem),
}))

export const ROUTE_TITLES: Record<string, string> = Object.fromEntries(
  ROUTE_META.map((meta) => [meta.path, meta.label]),
)

export function getVisibleNavGroups(permissions: readonly string[]): NavGroup[] {
  return NAV_GROUPS.map((group) => ({
    ...group,
    items: group.items.filter((item) => hasNavPermission(permissions, item)),
  })).filter((group) => group.items.length > 0)
}

function hasNavPermission(permissions: readonly string[], item: NavItem): boolean {
  if (item.requiredPermissions.length === 0) return true
  return hasPermission(permissions, item.requiredPermissions)
}

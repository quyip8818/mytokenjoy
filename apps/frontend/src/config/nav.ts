import type { LucideIcon } from 'lucide-react'
import { type PermissionKey, hasPermission } from '@/lib/permissions'
import { ROUTE_META, ROUTES, getRouteMeta, type RoutePath } from '@/config/routes'

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

export const NAV_GROUP_LAYOUT: {
  group: string
  paths: RoutePath[]
  collapsed?: boolean
}[] = [
  {
    group: '组织',
    paths: [ROUTES.orgDataSource, ROUTES.orgStructure, ROUTES.orgRoles],
  },
  {
    group: '预算',
    paths: [ROUTES.budgetOverview, ROUTES.budgetAllocation, ROUTES.budgetAlerts],
  },
  {
    group: '模型管理',
    paths: [ROUTES.modelsList, ROUTES.modelsRouting],
  },
  {
    group: 'Key 中心',
    paths: [ROUTES.keysMine, ROUTES.keysApproval, ROUTES.keysPlatform, ROUTES.keysProvider],
  },
  {
    group: '数据中心',
    collapsed: true,
    paths: [ROUTES.dashboardCost, ROUTES.dashboardUsage],
  },
  {
    group: '审计',
    collapsed: true,
    paths: [ROUTES.auditOperations, ROUTES.auditCalls],
  },
]

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
    collapsed: group.collapsed && !group.items.some((item) => hasNavPermission(permissions, item)),
    items: group.items.filter((item) => hasNavPermission(permissions, item)),
  })).filter((group) => group.items.length > 0)
}

function hasNavPermission(permissions: readonly string[], item: NavItem): boolean {
  return hasPermission(permissions, item.requiredPermissions)
}

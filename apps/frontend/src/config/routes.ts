import type { ComponentType } from 'react'
import {
  Activity,
  BarChart3,
  Building2,
  CheckCircle2,
  Cpu,
  CreditCard,
  Database,
  GitBranch,
  Globe,
  Key,
  PieChart,
  ScrollText,
  Shield,
  ShieldAlert,
  TrendingUp,
  Wallet,
  type LucideIcon,
} from 'lucide-react'
import { PERMISSION, type PermissionKey } from '@/lib/permission-keys'

type LazyPageModule = { default: ComponentType }

export interface RouteDefinition {
  key: string
  path: string
  label: string
  icon: LucideIcon
  requiredPermissions: readonly PermissionKey[]
  badgeKey?: 'approvalPending'
  lazy: () => Promise<LazyPageModule>
  navGroup: string
  navGroupCollapsed?: boolean
}

export const ROUTE_DEFINITIONS = [
  {
    key: 'orgDataSource',
    path: '/org/data-source',
    label: '数据源',
    icon: Database,
    requiredPermissions: [PERMISSION.ORG_DATASOURCE],
    lazy: () => import('@/routes/org/data-source'),
    navGroup: '组织',
  },
  {
    key: 'orgStructure',
    path: '/org/structure',
    label: '组织架构',
    icon: Building2,
    requiredPermissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS],
    lazy: () => import('@/routes/org/structure'),
    navGroup: '组织',
  },
  {
    key: 'orgRoles',
    path: '/org/roles',
    label: '角色管理',
    icon: Shield,
    requiredPermissions: [PERMISSION.ORG_ROLES],
    lazy: () => import('@/routes/org/roles'),
    navGroup: '组织',
  },
  {
    key: 'budgetOverview',
    path: '/budget/overview',
    label: '预算总览',
    icon: Wallet,
    requiredPermissions: [PERMISSION.BUDGET_READ],
    lazy: () => import('@/routes/budget/overview'),
    navGroup: '预算',
  },
  {
    key: 'budgetAllocation',
    path: '/budget/allocation',
    label: '预算分配',
    icon: PieChart,
    requiredPermissions: [PERMISSION.BUDGET_READ],
    lazy: () => import('@/routes/budget/allocation'),
    navGroup: '预算',
  },
  {
    key: 'budgetAlerts',
    path: '/budget/alerts',
    label: '超限策略',
    icon: ShieldAlert,
    requiredPermissions: [PERMISSION.BUDGET_POLICY],
    lazy: () => import('@/routes/budget/alerts'),
    navGroup: '预算',
  },
  {
    key: 'modelsList',
    path: '/models/list',
    label: '模型列表',
    icon: Cpu,
    requiredPermissions: [PERMISSION.MODEL_MANAGE],
    lazy: () => import('@/routes/models/list'),
    navGroup: '模型管理',
  },
  {
    key: 'modelsRouting',
    path: '/models/routing',
    label: '模型白名单',
    icon: GitBranch,
    requiredPermissions: [PERMISSION.MODEL_WHITELIST],
    lazy: () => import('@/routes/models/routing'),
    navGroup: '模型管理',
  },
  {
    key: 'keysMine',
    path: '/keys/mine',
    label: '我的 Key',
    icon: CreditCard,
    requiredPermissions: [PERMISSION.SELF_KEYS],
    lazy: () => import('@/routes/keys/mine'),
    navGroup: 'Key 中心',
  },
  {
    key: 'keysApproval',
    path: '/keys/approval',
    label: '审批中心',
    icon: CheckCircle2,
    requiredPermissions: [PERMISSION.BUDGET_APPROVE, PERMISSION.SELF_APPROVAL],
    badgeKey: 'approvalPending' as const,
    lazy: () => import('@/routes/keys/approval'),
    navGroup: 'Key 中心',
  },
  {
    key: 'keysPlatform',
    path: '/keys/platform',
    label: 'Key 管理',
    icon: Globe,
    requiredPermissions: [PERMISSION.KEYS_ADMIN],
    lazy: () => import('@/routes/keys/platform'),
    navGroup: 'Key 中心',
  },
  {
    key: 'keysProvider',
    path: '/keys/provider',
    label: '供应商 Key',
    icon: Key,
    requiredPermissions: [PERMISSION.KEYS_PROVIDER],
    lazy: () => import('@/routes/keys/provider'),
    navGroup: 'Key 中心',
  },
  {
    key: 'dashboardCost',
    path: '/dashboard/cost',
    label: '成本看板',
    icon: BarChart3,
    requiredPermissions: [PERMISSION.DASHBOARD_COST],
    lazy: () => import('@/routes/dashboard/cost'),
    navGroup: '数据中心',
    navGroupCollapsed: true,
  },
  {
    key: 'dashboardUsage',
    path: '/dashboard/usage',
    label: '用量分析',
    icon: TrendingUp,
    requiredPermissions: [PERMISSION.DASHBOARD_USAGE],
    lazy: () => import('@/routes/dashboard/usage'),
    navGroup: '数据中心',
  },
  {
    key: 'auditOperations',
    path: '/audit/operations',
    label: '操作审计',
    icon: ScrollText,
    requiredPermissions: [PERMISSION.AUDIT_READ],
    lazy: () => import('@/routes/audit/operations'),
    navGroup: '审计',
    navGroupCollapsed: true,
  },
  {
    key: 'auditCalls',
    path: '/audit/calls',
    label: '调用日志',
    icon: Activity,
    requiredPermissions: [PERMISSION.AUDIT_READ],
    lazy: () => import('@/routes/audit/calls'),
    navGroup: '审计',
  },
] as const satisfies readonly RouteDefinition[]

export type RouteKey = (typeof ROUTE_DEFINITIONS)[number]['key']

const routeEntries = ROUTE_DEFINITIONS.map((definition) => [definition.key, definition.path] as const)

export const ROUTES = {
  home: '/',
  ...Object.fromEntries(routeEntries),
} as { home: '/' } & Record<RouteKey, string>

export type RoutePath = '/' | (typeof ROUTE_DEFINITIONS)[number]['path']

export interface RouteMeta {
  path: RoutePath
  label: string
  icon: LucideIcon
  requiredPermissions: readonly PermissionKey[]
  badgeKey?: 'approvalPending'
}

export const ROUTE_META: RouteMeta[] = ROUTE_DEFINITIONS.map((definition) => ({
  path: definition.path as RoutePath,
  label: definition.label,
  icon: definition.icon,
  requiredPermissions: definition.requiredPermissions,
  ...('badgeKey' in definition && definition.badgeKey ? { badgeKey: definition.badgeKey } : {}),
}))

export interface NavGroupLayoutEntry {
  group: string
  paths: RoutePath[]
  collapsed?: boolean
}

export const NAV_GROUP_LAYOUT: NavGroupLayoutEntry[] = (() => {
  const groups: NavGroupLayoutEntry[] = []
  for (const definition of ROUTE_DEFINITIONS) {
    let group = groups.find((entry) => entry.group === definition.navGroup)
    if (!group) {
      group = {
        group: definition.navGroup,
        paths: [],
        ...('navGroupCollapsed' in definition && definition.navGroupCollapsed
          ? { collapsed: definition.navGroupCollapsed }
          : {}),
      }
      groups.push(group)
    }
    group.paths.push(definition.path as RoutePath)
  }
  return groups
})()

export const HOME_ROUTE_KEYS = [
  'orgDataSource',
  'keysApproval',
  'budgetOverview',
  'keysMine',
  'auditOperations',
  'dashboardCost',
] as const satisfies readonly RouteKey[]

export const HOME_PATH_CANDIDATES = HOME_ROUTE_KEYS.map((key) => ROUTES[key]) as RoutePath[]

export function getRouteMeta(path: RoutePath): RouteMeta {
  const meta = ROUTE_META.find((entry) => entry.path === path)
  if (!meta) {
    throw new Error(`Missing ROUTE_META for path: ${path}`)
  }
  return meta
}

export function routePermissions(path: RoutePath): PermissionKey[] {
  return [...getRouteMeta(path).requiredPermissions]
}

export interface AppRouteEntry {
  path: RoutePath
  lazy: () => Promise<LazyPageModule>
}

export const APP_ROUTES: AppRouteEntry[] = ROUTE_DEFINITIONS.map(({ path, lazy }) => ({
  path: path as RoutePath,
  lazy,
}))

export function toRouterPath(route: RoutePath): string {
  return route === ROUTES.home ? '' : route.slice(1)
}

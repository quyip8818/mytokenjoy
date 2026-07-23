import type { ComponentType } from 'react'
import {
  Activity,
  BarChart3,
  Building2,
  CheckCircle2,
  Cpu,
  Database,
  FileText,
  GitBranch,
  Globe,
  Key,
  ScrollText,
  Settings,
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
  lazy: () => Promise<LazyPageModule>
  /**
   * Route is accessible if the user holds ANY of the listed permissions (OR semantics).
   * An empty or absent array means no permission is required (visible to all).
   */
  requiredPermissions?: readonly PermissionKey[]
  badgeKey?: 'approvalPending'
  navGroup?: string
  navGroupCollapsed?: boolean
}

const ROUTE_DEFINITIONS_INTERNAL = [
  // ─── 数据看板（默认展开）───
  {
    key: 'dashboardCost',
    path: '/dashboard/cost',
    label: '成本看板',
    icon: BarChart3,
    requiredPermissions: [PERMISSION.DASHBOARD_COST],
    lazy: () => import('@/routes/dashboard/cost'),
    navGroup: '数据看板',
  },
  {
    key: 'dashboardUsage',
    path: '/dashboard/usage',
    label: '用量分析',
    icon: TrendingUp,
    requiredPermissions: [PERMISSION.DASHBOARD_USAGE],
    lazy: () => import('@/routes/dashboard/usage'),
    navGroup: '数据看板',
  },
  // ─── 凭证管理（默认展开）───
  {
    key: 'keysPlatform',
    path: '/keys/platform',
    label: 'Key 管理',
    icon: Globe,
    requiredPermissions: [PERMISSION.KEYS_ADMIN, PERMISSION.KEYS_READ],
    lazy: () => import('@/routes/keys/platform'),
    navGroup: '凭证管理',
  },
  {
    key: 'keysApproval',
    path: '/approvals',
    label: '审批中心',
    icon: CheckCircle2,
    requiredPermissions: [PERMISSION.BUDGET_APPROVE, PERMISSION.SELF_APPROVAL],
    badgeKey: 'approvalPending' as const,
    lazy: () => import('@/routes/approval/index'),
    navGroup: '凭证管理',
  },
  {
    key: 'keysProvider',
    path: '/keys/provider',
    label: '供应商 Key',
    icon: Key,
    requiredPermissions: [PERMISSION.KEYS_PROVIDER, PERMISSION.KEYS_READ],
    lazy: () => import('@/routes/keys/provider'),
    navGroup: '凭证管理',
  },
  // ─── 模型管理（默认展开）───
  {
    key: 'modelsList',
    path: '/models/list',
    label: '模型列表',
    icon: Cpu,
    requiredPermissions: [PERMISSION.MODEL_MANAGE, PERMISSION.MODEL_READ],
    lazy: () => import('@/routes/models/list'),
    navGroup: '模型管理',
  },
  {
    key: 'modelsRouting',
    path: '/models/routing',
    label: '模型配置',
    icon: GitBranch,
    requiredPermissions: [PERMISSION.MODEL_WHITELIST],
    lazy: () => import('@/routes/models/routing'),
    navGroup: '模型管理',
  },
  // ─── 预算与财务（默认折叠）───
  {
    key: 'budget',
    path: '/budget',
    label: '预算管理',
    icon: Wallet,
    requiredPermissions: [PERMISSION.BUDGET_READ],
    lazy: () => import('@/routes/budget'),
    navGroup: '预算与财务',
    navGroupCollapsed: true,
  },
  {
    key: 'budgetAlerts',
    path: '/budget/alerts',
    label: '预警规则',
    icon: ShieldAlert,
    requiredPermissions: [PERMISSION.BUDGET_POLICY],
    lazy: () => import('@/routes/budget/alerts'),
    navGroup: '预算与财务',
  },
  {
    key: 'wallet',
    path: '/wallet',
    label: '钱包管理',
    icon: Wallet,
    requiredPermissions: [PERMISSION.BILLING_READ],
    lazy: () => import('@/routes/wallet'),
    navGroup: '预算与财务',
  },
  // ─── 组织与权限（默认折叠）───
  {
    key: 'orgDataSource',
    path: '/org/data-source',
    label: '数据源',
    icon: Database,
    requiredPermissions: [PERMISSION.ORG_DATASOURCE],
    lazy: () => import('@/routes/org/data-source'),
    navGroup: '组织与权限',
    navGroupCollapsed: true,
  },
  {
    key: 'orgStructure',
    path: '/org/structure',
    label: '组织架构',
    icon: Building2,
    requiredPermissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS, PERMISSION.ORG_READ],
    lazy: () => import('@/routes/org/structure'),
    navGroup: '组织与权限',
  },
  {
    key: 'orgRoles',
    path: '/org/roles',
    label: '角色管理',
    icon: Shield,
    requiredPermissions: [PERMISSION.ORG_ROLES, PERMISSION.ORG_READ],
    lazy: () => import('@/routes/org/roles'),
    navGroup: '组织与权限',
  },
  // ─── 审计（默认折叠）───
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
  // ─── 我的（默认折叠）───
  {
    key: 'myKeys',
    path: '/me/keys',
    label: '我的 Key',
    icon: Key,
    lazy: () => import('@/routes/member/keys'),
    navGroup: '我的',
    navGroupCollapsed: true,
  },
  {
    key: 'myUsage',
    path: '/me/usage',
    label: '我的用量',
    icon: FileText,
    lazy: () => import('@/routes/member/usage'),
    navGroup: '我的',
  },
  {
    key: 'mySettings',
    path: '/me/settings',
    label: '设置',
    icon: Settings,
    lazy: () => import('@/routes/member/settings'),
    navGroup: '我的',
  },
] as const satisfies readonly RouteDefinition[]

export const ROUTE_DEFINITIONS = ROUTE_DEFINITIONS_INTERNAL

export type RouteKey = (typeof ROUTE_DEFINITIONS)[number]['key']

const routeEntries = ROUTE_DEFINITIONS.map(
  (definition) => [definition.key, definition.path] as const,
)

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
  requiredPermissions: definition.requiredPermissions ?? [],
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
    if (!definition.navGroup) continue
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
  'dashboardCost',
  'keysApproval',
  'budget',
  'myKeys',
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

const LAZY_IMPORT_PATTERN = /import\(['"](@\/routes\/[^'"]+)['"]\)/

export function validateRouteDefinitions(): void {
  const keys = ROUTE_DEFINITIONS.map((definition) => definition.key)
  const paths = ROUTE_DEFINITIONS.map((definition) => definition.path)

  if (new Set(keys).size !== keys.length) {
    throw new Error('ROUTE_DEFINITIONS contains duplicate keys')
  }

  if (new Set(paths).size !== paths.length) {
    throw new Error('ROUTE_DEFINITIONS contains duplicate paths')
  }

  const lazyImportMatches = ROUTE_DEFINITIONS.map((definition) =>
    definition.lazy.toString().match(LAZY_IMPORT_PATTERN),
  )

  if (lazyImportMatches.some((match) => !match)) {
    throw new Error('ROUTE_DEFINITIONS lazy imports must use import("@/routes/...") syntax')
  }
}

export function getRouteLazyImportPaths(): string[] {
  return ROUTE_DEFINITIONS.map((definition) => {
    const match = definition.lazy.toString().match(LAZY_IMPORT_PATTERN)
    if (!match?.[1]) {
      throw new Error(`Unable to resolve lazy import for route ${definition.key}`)
    }
    return match[1]
  })
}

export function toRouterPath(route: RoutePath): string {
  return route === ROUTES.home ? '' : route.slice(1)
}

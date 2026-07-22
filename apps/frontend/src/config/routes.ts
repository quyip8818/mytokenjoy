import type { ComponentType } from 'react'
import {
  Activity,
  BarChart3,
  Bell,
  Building2,
  CheckCircle2,
  Cpu,
  CreditCard,
  Database,
  FileText,
  GitBranch,
  Globe,
  Key,
  LayoutDashboard,
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
  audience: 'admin' | 'member'
  lazy: () => Promise<LazyPageModule>
  /**
   * Route is accessible if the user holds ANY of the listed permissions (OR semantics).
   * An empty or absent array means no permission is required.
   */
  requiredPermissions?: readonly PermissionKey[]
  badgeKey?: 'approvalPending'
  navGroup?: string
  navGroupCollapsed?: boolean
  navEnd?: boolean
}

const ADMIN_ROUTE_DEFINITIONS = [
  {
    key: 'orgDataSource',
    path: '/org/data-source',
    label: '数据源',
    icon: Database,
    audience: 'admin',
    requiredPermissions: [PERMISSION.ORG_DATASOURCE],
    lazy: () => import('@/routes/org/data-source'),
    navGroup: '组织',
  },
  {
    key: 'orgStructure',
    path: '/org/structure',
    label: '组织架构',
    icon: Building2,
    audience: 'admin',
    requiredPermissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS, PERMISSION.ORG_READ],
    lazy: () => import('@/routes/org/structure'),
    navGroup: '组织',
  },
  {
    key: 'orgRoles',
    path: '/org/roles',
    label: '角色管理',
    icon: Shield,
    audience: 'admin',
    requiredPermissions: [PERMISSION.ORG_ROLES, PERMISSION.ORG_READ],
    lazy: () => import('@/routes/org/roles'),
    navGroup: '组织',
  },
  {
    key: 'budget',
    path: '/budget',
    label: '预算管理',
    icon: Wallet,
    audience: 'admin',
    requiredPermissions: [PERMISSION.BUDGET_READ],
    lazy: () => import('@/routes/budget'),
    navGroup: '预算',
  },
  {
    key: 'budgetAlerts',
    path: '/budget/alerts',
    label: '预警规则',
    icon: ShieldAlert,
    audience: 'admin',
    requiredPermissions: [PERMISSION.BUDGET_POLICY],
    lazy: () => import('@/routes/budget/alerts'),
    navGroup: '预算',
  },
  {
    key: 'modelsList',
    path: '/models/list',
    label: '模型列表',
    icon: Cpu,
    audience: 'admin',
    requiredPermissions: [PERMISSION.MODEL_MANAGE, PERMISSION.MODEL_READ],
    lazy: () => import('@/routes/models/list'),
    navGroup: '模型管理',
  },
  {
    key: 'modelsRouting',
    path: '/models/routing',
    label: '模型配置',
    icon: GitBranch,
    audience: 'admin',
    requiredPermissions: [PERMISSION.MODEL_WHITELIST],
    lazy: () => import('@/routes/models/routing'),
    navGroup: '模型管理',
  },
  {
    key: 'keysMine',
    path: '/keys/mine',
    label: '我的 Key',
    icon: CreditCard,
    audience: 'admin',
    requiredPermissions: [PERMISSION.SELF_KEYS],
    lazy: () => import('@/routes/keys/mine'),
    navGroup: 'Key 中心',
  },
  {
    key: 'keysApproval',
    path: '/approvals',
    label: '审批中心',
    icon: CheckCircle2,
    audience: 'admin',
    requiredPermissions: [PERMISSION.BUDGET_APPROVE, PERMISSION.SELF_APPROVAL],
    badgeKey: 'approvalPending' as const,
    lazy: () => import('@/routes/approval/index'),
    navGroup: 'Key 中心',
  },
  {
    key: 'keysPlatform',
    path: '/keys/platform',
    label: 'Key 管理',
    icon: Globe,
    audience: 'admin',
    requiredPermissions: [PERMISSION.KEYS_ADMIN, PERMISSION.KEYS_READ],
    lazy: () => import('@/routes/keys/platform'),
    navGroup: 'Key 中心',
  },
  {
    key: 'keysProvider',
    path: '/keys/provider',
    label: '供应商 Key',
    icon: Key,
    audience: 'admin',
    requiredPermissions: [PERMISSION.KEYS_PROVIDER, PERMISSION.KEYS_READ],
    lazy: () => import('@/routes/keys/provider'),
    navGroup: 'Key 中心',
  },
  {
    key: 'dashboardCost',
    path: '/dashboard/cost',
    label: '成本看板',
    icon: BarChart3,
    audience: 'admin',
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
    audience: 'admin',
    requiredPermissions: [PERMISSION.DASHBOARD_USAGE],
    lazy: () => import('@/routes/dashboard/usage'),
    navGroup: '数据中心',
  },
  {
    key: 'wallet',
    path: '/wallet',
    label: '钱包管理',
    icon: Wallet,
    audience: 'admin',
    requiredPermissions: [PERMISSION.BILLING_READ],
    lazy: () => import('@/routes/wallet'),
    navGroup: '财务管理',
  },
  {
    key: 'auditOperations',
    path: '/audit/operations',
    label: '操作审计',
    icon: ScrollText,
    audience: 'admin',
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
    audience: 'admin',
    requiredPermissions: [PERMISSION.AUDIT_READ],
    lazy: () => import('@/routes/audit/calls'),
    navGroup: '审计',
  },
] as const satisfies readonly RouteDefinition[]

const MEMBER_ROUTE_DEFINITIONS_INTERNAL = [
  {
    key: 'memberDashboard',
    path: '/me',
    label: '工作台',
    icon: LayoutDashboard,
    audience: 'member',
    lazy: () => import('@/routes/member'),
    navEnd: true,
  },
  {
    key: 'memberKeys',
    path: '/me/keys',
    label: '我的 Key',
    icon: Key,
    audience: 'member',
    lazy: () => import('@/routes/member/keys'),
  },
  {
    key: 'memberCallLogs',
    path: '/me/call-logs',
    label: '使用记录',
    icon: FileText,
    audience: 'member',
    lazy: () => import('@/routes/member/call-logs'),
  },
  {
    key: 'memberNotifications',
    path: '/me/notifications',
    label: '通知偏好',
    icon: Bell,
    audience: 'member',
    lazy: () => import('@/routes/member/notifications'),
  },
  {
    key: 'memberAccount',
    path: '/me/account',
    label: '账户设置',
    icon: Settings,
    audience: 'member',
    lazy: () => import('@/routes/member/account'),
  },
  {
    key: 'memberLoginActivity',
    path: '/me/login-activity',
    label: '登录活动',
    icon: Activity,
    audience: 'member',
    lazy: () => import('@/routes/member/login-activity'),
  },
] as const satisfies readonly RouteDefinition[]

export const ALL_ROUTE_DEFINITIONS = [
  ...ADMIN_ROUTE_DEFINITIONS,
  ...MEMBER_ROUTE_DEFINITIONS_INTERNAL,
] as const satisfies readonly RouteDefinition[]

export const ROUTE_DEFINITIONS = ADMIN_ROUTE_DEFINITIONS
export const MEMBER_ROUTE_DEFINITIONS: readonly RouteDefinition[] =
  MEMBER_ROUTE_DEFINITIONS_INTERNAL

export type MemberRoutePath = (typeof MEMBER_ROUTE_DEFINITIONS)[number]['path']

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
  'orgDataSource',
  'keysApproval',
  'budget',
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

const LAZY_IMPORT_PATTERN = /import\(['"](@\/routes\/[^'"]+)['"]\)/

export function validateRouteDefinitions(): void {
  const keys = ALL_ROUTE_DEFINITIONS.map((definition) => definition.key)
  const paths = ALL_ROUTE_DEFINITIONS.map((definition) => definition.path)

  if (new Set(keys).size !== keys.length) {
    throw new Error('ALL_ROUTE_DEFINITIONS contains duplicate keys')
  }

  if (new Set(paths).size !== paths.length) {
    throw new Error('ALL_ROUTE_DEFINITIONS contains duplicate paths')
  }

  const lazyImportMatches = ALL_ROUTE_DEFINITIONS.map((definition) =>
    definition.lazy.toString().match(LAZY_IMPORT_PATTERN),
  )

  if (lazyImportMatches.some((match) => !match)) {
    throw new Error('ALL_ROUTE_DEFINITIONS lazy imports must use import("@/routes/...") syntax')
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

export function getMemberRouteLazyImportPaths(): string[] {
  return MEMBER_ROUTE_DEFINITIONS.map((definition) => {
    const match = definition.lazy.toString().match(LAZY_IMPORT_PATTERN)
    if (!match?.[1]) {
      throw new Error(`Unable to resolve lazy import for member route ${definition.key}`)
    }
    return match[1]
  })
}

export function toMemberRouterPath(path: MemberRoutePath): string {
  return path.slice(1)
}

export function toRouterPath(route: RoutePath): string {
  return route === ROUTES.home ? '' : route.slice(1)
}

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

export const ROUTES = {
  home: '/',
  dashboardCost: '/dashboard/cost',
  dashboardUsage: '/dashboard/usage',
  orgDataSource: '/org/data-source',
  orgStructure: '/org/structure',
  orgRoles: '/org/roles',
  budgetOverview: '/budget/overview',
  budgetAllocation: '/budget/allocation',
  budgetAlerts: '/budget/alerts',
  keysProvider: '/keys/provider',
  keysPlatform: '/keys/platform',
  keysMine: '/keys/mine',
  keysApproval: '/keys/approval',
  modelsList: '/models/list',
  modelsRouting: '/models/routing',
  auditOperations: '/audit/operations',
  auditCalls: '/audit/calls',
} as const

export type RoutePath = (typeof ROUTES)[keyof typeof ROUTES]

export interface RouteMeta {
  path: RoutePath
  label: string
  icon: LucideIcon
  requiredPermissions: readonly PermissionKey[]
  badgeKey?: 'approvalPending'
}

export const ROUTE_META: RouteMeta[] = [
  {
    path: ROUTES.orgDataSource,
    label: '数据源',
    icon: Database,
    requiredPermissions: [PERMISSION.ORG_DATASOURCE],
  },
  {
    path: ROUTES.orgStructure,
    label: '组织架构',
    icon: Building2,
    requiredPermissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS],
  },
  {
    path: ROUTES.orgRoles,
    label: '角色管理',
    icon: Shield,
    requiredPermissions: [PERMISSION.ORG_ROLES],
  },
  {
    path: ROUTES.budgetOverview,
    label: '预算总览',
    icon: Wallet,
    requiredPermissions: [PERMISSION.BUDGET_READ],
  },
  {
    path: ROUTES.budgetAllocation,
    label: '预算分配',
    icon: PieChart,
    requiredPermissions: [PERMISSION.BUDGET_READ],
  },
  {
    path: ROUTES.budgetAlerts,
    label: '超限策略',
    icon: ShieldAlert,
    requiredPermissions: [PERMISSION.BUDGET_POLICY],
  },
  {
    path: ROUTES.modelsList,
    label: '模型列表',
    icon: Cpu,
    requiredPermissions: [PERMISSION.MODEL_MANAGE],
  },
  {
    path: ROUTES.modelsRouting,
    label: '模型白名单',
    icon: GitBranch,
    requiredPermissions: [PERMISSION.MODEL_WHITELIST],
  },
  {
    path: ROUTES.keysMine,
    label: '我的 Key',
    icon: CreditCard,
    requiredPermissions: [PERMISSION.SELF_KEYS],
  },
  {
    path: ROUTES.keysApproval,
    label: '审批中心',
    icon: CheckCircle2,
    requiredPermissions: [PERMISSION.BUDGET_APPROVE, PERMISSION.SELF_APPROVAL],
    badgeKey: 'approvalPending',
  },
  {
    path: ROUTES.keysPlatform,
    label: 'Key 管理',
    icon: Globe,
    requiredPermissions: [PERMISSION.KEYS_ADMIN],
  },
  {
    path: ROUTES.keysProvider,
    label: '供应商 Key',
    icon: Key,
    requiredPermissions: [PERMISSION.KEYS_PROVIDER],
  },
  {
    path: ROUTES.dashboardCost,
    label: '成本看板',
    icon: BarChart3,
    requiredPermissions: [PERMISSION.DASHBOARD_COST],
  },
  {
    path: ROUTES.dashboardUsage,
    label: '用量分析',
    icon: TrendingUp,
    requiredPermissions: [PERMISSION.DASHBOARD_USAGE],
  },
  {
    path: ROUTES.auditOperations,
    label: '操作审计',
    icon: ScrollText,
    requiredPermissions: [PERMISSION.AUDIT_READ],
  },
  {
    path: ROUTES.auditCalls,
    label: '调用日志',
    icon: Activity,
    requiredPermissions: [PERMISSION.AUDIT_READ],
  },
]

export const HOME_PATH_CANDIDATES: RoutePath[] = [
  ROUTES.orgDataSource,
  ROUTES.keysApproval,
  ROUTES.budgetOverview,
  ROUTES.keysMine,
  ROUTES.auditOperations,
  ROUTES.dashboardCost,
]

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

type LazyPageModule = { default: ComponentType }

export interface AppRouteEntry {
  path: RoutePath
  lazy: () => Promise<LazyPageModule>
}

export const APP_ROUTES: AppRouteEntry[] = [
  { path: ROUTES.dashboardCost, lazy: () => import('@/routes/dashboard/cost') },
  { path: ROUTES.dashboardUsage, lazy: () => import('@/routes/dashboard/usage') },
  { path: ROUTES.orgDataSource, lazy: () => import('@/routes/org/data-source') },
  { path: ROUTES.orgStructure, lazy: () => import('@/routes/org/structure') },
  { path: ROUTES.orgRoles, lazy: () => import('@/routes/org/roles') },
  { path: ROUTES.budgetOverview, lazy: () => import('@/routes/budget/overview') },
  { path: ROUTES.budgetAllocation, lazy: () => import('@/routes/budget/allocation') },
  { path: ROUTES.budgetAlerts, lazy: () => import('@/routes/budget/alerts') },
  { path: ROUTES.keysProvider, lazy: () => import('@/routes/keys/provider') },
  { path: ROUTES.keysPlatform, lazy: () => import('@/routes/keys/platform') },
  { path: ROUTES.keysMine, lazy: () => import('@/routes/keys/mine') },
  { path: ROUTES.keysApproval, lazy: () => import('@/routes/keys/approval') },
  { path: ROUTES.modelsList, lazy: () => import('@/routes/models/list') },
  { path: ROUTES.modelsRouting, lazy: () => import('@/routes/models/routing') },
  { path: ROUTES.auditOperations, lazy: () => import('@/routes/audit/operations') },
  { path: ROUTES.auditCalls, lazy: () => import('@/routes/audit/calls') },
]

export function toRouterPath(route: RoutePath): string {
  return route === ROUTES.home ? '' : route.slice(1)
}

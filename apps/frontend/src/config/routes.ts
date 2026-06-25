import type { ComponentType } from 'react'
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
  requiredPermissions: readonly PermissionKey[]
}

export const ROUTE_META: RouteMeta[] = [
  { path: ROUTES.orgDataSource, requiredPermissions: [PERMISSION.ORG_DATASOURCE] },
  {
    path: ROUTES.orgStructure,
    requiredPermissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS],
  },
  { path: ROUTES.orgRoles, requiredPermissions: [PERMISSION.ORG_ROLES] },
  { path: ROUTES.budgetOverview, requiredPermissions: [PERMISSION.BUDGET_READ] },
  { path: ROUTES.budgetAllocation, requiredPermissions: [PERMISSION.BUDGET_READ] },
  { path: ROUTES.budgetAlerts, requiredPermissions: [PERMISSION.BUDGET_POLICY] },
  { path: ROUTES.modelsList, requiredPermissions: [PERMISSION.MODEL_MANAGE] },
  { path: ROUTES.modelsRouting, requiredPermissions: [PERMISSION.MODEL_WHITELIST] },
  { path: ROUTES.keysMine, requiredPermissions: [PERMISSION.SELF_KEYS] },
  {
    path: ROUTES.keysApproval,
    requiredPermissions: [PERMISSION.BUDGET_APPROVE, PERMISSION.SELF_APPROVAL],
  },
  { path: ROUTES.keysPlatform, requiredPermissions: [PERMISSION.KEYS_ADMIN] },
  { path: ROUTES.keysProvider, requiredPermissions: [PERMISSION.KEYS_PROVIDER] },
  { path: ROUTES.dashboardCost, requiredPermissions: [PERMISSION.DASHBOARD_COST] },
  { path: ROUTES.dashboardUsage, requiredPermissions: [PERMISSION.DASHBOARD_USAGE] },
  { path: ROUTES.auditOperations, requiredPermissions: [PERMISSION.AUDIT_READ] },
  { path: ROUTES.auditCalls, requiredPermissions: [PERMISSION.AUDIT_READ] },
]

export const HOME_PATH_CANDIDATES: RouteMeta[] = [
  { path: ROUTES.orgDataSource, requiredPermissions: [PERMISSION.ORG_DATASOURCE] },
  { path: ROUTES.keysApproval, requiredPermissions: [PERMISSION.BUDGET_APPROVE] },
  { path: ROUTES.budgetOverview, requiredPermissions: [PERMISSION.BUDGET_READ] },
  { path: ROUTES.keysMine, requiredPermissions: [PERMISSION.SELF_KEYS] },
  { path: ROUTES.auditOperations, requiredPermissions: [PERMISSION.AUDIT_READ] },
  { path: ROUTES.dashboardCost, requiredPermissions: [PERMISSION.DASHBOARD_COST] },
]

export function routePermissions(path: RoutePath): PermissionKey[] {
  const meta = ROUTE_META.find((entry) => entry.path === path)
  if (!meta) {
    throw new Error(`Missing ROUTE_META for path: ${path}`)
  }
  return [...meta.requiredPermissions]
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

import type { Member, Role } from '@/api/types'
import { HOME_PATH_CANDIDATES, ROUTE_META, ROUTES } from '@/config/routes'
import {
  ROLE_API_CALLER,
  ROLE_AUDITOR,
  ROLE_MEMBER,
  ROLE_ORG_ADMIN,
  ROLE_SUPER_ADMIN,
} from '@/lib/role-constants'

export const PERMISSION = {
  ORG_DATASOURCE: 'org:datasource',
  ORG_STRUCTURE: 'org:structure',
  ORG_ROLES: 'org:roles',
  ORG_MEMBERS: 'org:members',
  BUDGET_READ: 'budget:read',
  BUDGET_ALLOCATE: 'budget:allocate',
  BUDGET_APPROVE: 'budget:approve',
  BUDGET_POLICY: 'budget:policy',
  MODEL_MANAGE: 'model:manage',
  MODEL_WHITELIST: 'model:whitelist',
  KEYS_ADMIN: 'keys:admin',
  KEYS_PROVIDER: 'keys:provider',
  SELF_KEYS: 'self:keys',
  SELF_APPROVAL: 'self:approval',
  DASHBOARD_COST: 'dashboard:cost',
  DASHBOARD_USAGE: 'dashboard:usage',
  AUDIT_READ: 'audit:read',
  API_CALL: 'api:call',
} as const

export type PermissionKey = (typeof PERMISSION)[keyof typeof PERMISSION]

export const ALL_PERMISSIONS: PermissionKey[] = Object.values(PERMISSION)

export const PERMISSION_ID_MAP: Record<string, PermissionKey> = {
  'p-1': PERMISSION.ORG_STRUCTURE,
  'p-2': PERMISSION.ORG_MEMBERS,
  'p-3': PERMISSION.ORG_ROLES,
  'p-4': PERMISSION.ORG_DATASOURCE,
  'p-5': PERMISSION.BUDGET_ALLOCATE,
  'p-6': PERMISSION.BUDGET_APPROVE,
  'p-7': PERMISSION.MODEL_WHITELIST,
  'p-8': PERMISSION.DASHBOARD_COST,
  'p-9': PERMISSION.DASHBOARD_USAGE,
  'p-10': PERMISSION.AUDIT_READ,
  'p-11': PERMISSION.API_CALL,
  'p-12': PERMISSION.BUDGET_READ,
  'p-13': PERMISSION.BUDGET_POLICY,
  'p-14': PERMISSION.MODEL_MANAGE,
  'p-15': PERMISSION.KEYS_ADMIN,
  'p-16': PERMISSION.KEYS_PROVIDER,
  'p-17': PERMISSION.SELF_KEYS,
  'p-18': PERMISSION.SELF_APPROVAL,
}

const BUDGET_WRITE_CAPABILITIES: PermissionKey[] = [
  PERMISSION.BUDGET_ALLOCATE,
  PERMISSION.BUDGET_APPROVE,
  PERMISSION.BUDGET_POLICY,
]

const PRESET_ROLE_CAPABILITIES: Record<string, PermissionKey[]> = {
  [ROLE_SUPER_ADMIN]: [...ALL_PERMISSIONS],
  [ROLE_ORG_ADMIN]: [...ALL_PERMISSIONS],
  [ROLE_MEMBER]: [PERMISSION.SELF_KEYS, PERMISSION.SELF_APPROVAL],
  [ROLE_AUDITOR]: [
    PERMISSION.AUDIT_READ,
    PERMISSION.DASHBOARD_COST,
    PERMISSION.DASHBOARD_USAGE,
    PERMISSION.SELF_APPROVAL,
  ],
  [ROLE_API_CALLER]: [PERMISSION.API_CALL],
}

const WRITE_CAPABILITIES: PermissionKey[] = [
  PERMISSION.ORG_DATASOURCE,
  PERMISSION.ORG_STRUCTURE,
  PERMISSION.ORG_ROLES,
  PERMISSION.ORG_MEMBERS,
  PERMISSION.BUDGET_ALLOCATE,
  PERMISSION.BUDGET_APPROVE,
  PERMISSION.BUDGET_POLICY,
  PERMISSION.MODEL_MANAGE,
  PERMISSION.MODEL_WHITELIST,
  PERMISSION.KEYS_ADMIN,
  PERMISSION.KEYS_PROVIDER,
]

function expandRoleDefinition(role: Role): PermissionKey[] {
  if (role.type === 'preset') {
    return PRESET_ROLE_CAPABILITIES[role.name] ?? []
  }

  const caps = new Set<PermissionKey>()
  for (const permId of role.permissions) {
    if (permId === '*') {
      ALL_PERMISSIONS.forEach((p) => caps.add(p))
      continue
    }
    const mapped = PERMISSION_ID_MAP[permId]
    if (mapped) {
      caps.add(mapped)
    } else if (ALL_PERMISSIONS.includes(permId as PermissionKey)) {
      caps.add(permId as PermissionKey)
    }
  }
  const expanded = [...caps]
  if (expanded.some((p) => BUDGET_WRITE_CAPABILITIES.includes(p))) {
    caps.add(PERMISSION.BUDGET_READ)
  }
  return [...caps]
}

export function resolveMemberPermissions(member: Member, roles: Role[]): PermissionKey[] {
  const caps = new Set<PermissionKey>()
  for (const roleName of member.roles) {
    const roleDef = roles.find((r) => r.name === roleName)
    if (!roleDef) continue
    expandRoleDefinition(roleDef).forEach((p) => caps.add(p))
  }
  return [...caps]
}

export function hasPermission(
  permissions: readonly string[],
  required: PermissionKey | PermissionKey[],
): boolean {
  const requiredList = Array.isArray(required) ? required : [required]
  return requiredList.some((p) => permissions.includes(p))
}

export function isReadOnlySession(permissions: readonly string[]): boolean {
  if (permissions.includes('*' as PermissionKey)) return false
  const hasWrite = WRITE_CAPABILITIES.some((p) => permissions.includes(p))
  return !hasWrite
}

export function canWriteSession(permissions: readonly string[]): boolean {
  return !isReadOnlySession(permissions)
}

export function getDefaultHomePath(permissions: readonly string[]): string {
  for (const { path, requiredPermissions } of HOME_PATH_CANDIDATES) {
    if (hasPermission(permissions, [...requiredPermissions])) return path
  }
  return ROUTES.dashboardCost
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

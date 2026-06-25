import { describe, expect, it } from 'vitest'
import type { Member, Role } from '@/api/types'
import { ROUTES } from '@/config/routes'
import { PERMISSION } from '@/lib/permission-keys'
import { ROLE_API_CALLER, ROLE_AUDITOR, ROLE_MEMBER, ROLE_SUPER_ADMIN } from '@/lib/role-constants'
import {
  canAccessRoute,
  getDefaultHomePath,
  hasPermission,
  isReadOnlySession,
  resolveMemberPermissions,
} from './permissions'

const roles: Role[] = [
  { id: 'r1', name: ROLE_SUPER_ADMIN, type: 'preset', permissions: [], memberCount: 1 },
  { id: 'r2', name: ROLE_MEMBER, type: 'preset', permissions: [], memberCount: 1 },
  { id: 'r3', name: ROLE_AUDITOR, type: 'preset', permissions: [], memberCount: 1 },
  { id: 'r4', name: ROLE_API_CALLER, type: 'preset', permissions: [], memberCount: 1 },
]

function memberWithRoles(roleNames: string[]): Member {
  return {
    id: 'm1',
    name: 'Test',
    phone: '13800000000',
    email: 'test@test.com',
    departmentId: 'd1',
    departmentName: '总部',
    status: 'active',
    roles: roleNames,
    source: 'manual',
  }
}

describe('hasPermission', () => {
  it('returns true when any required permission matches', () => {
    expect(hasPermission([PERMISSION.ORG_STRUCTURE], PERMISSION.ORG_STRUCTURE)).toBe(true)
    expect(
      hasPermission([PERMISSION.SELF_KEYS], [PERMISSION.ORG_STRUCTURE, PERMISSION.SELF_KEYS]),
    ).toBe(true)
  })

  it('returns false when no required permission matches', () => {
    expect(hasPermission([PERMISSION.SELF_KEYS], PERMISSION.ORG_STRUCTURE)).toBe(false)
  })
})

describe('isReadOnlySession', () => {
  it('returns false for sessions with write capabilities', () => {
    expect(isReadOnlySession([PERMISSION.ORG_STRUCTURE])).toBe(false)
  })

  it('returns true for read-only auditor permissions', () => {
    const perms = resolveMemberPermissions(memberWithRoles([ROLE_AUDITOR]), roles)
    expect(isReadOnlySession(perms)).toBe(true)
  })
})

describe('resolveMemberPermissions', () => {
  it('expands preset super admin to all permissions', () => {
    const perms = resolveMemberPermissions(memberWithRoles([ROLE_SUPER_ADMIN]), roles)
    expect(perms).toContain(PERMISSION.ORG_STRUCTURE)
    expect(perms).toContain(PERMISSION.BUDGET_READ)
  })

  it('expands member role to self-service permissions', () => {
    const perms = resolveMemberPermissions(memberWithRoles([ROLE_MEMBER]), roles)
    expect(perms).toEqual([PERMISSION.SELF_KEYS, PERMISSION.SELF_APPROVAL])
  })
})

describe('canAccessRoute', () => {
  it('allows access when user has required route permission', () => {
    expect(canAccessRoute(ROUTES.orgStructure, [PERMISSION.ORG_STRUCTURE])).toBe(true)
  })

  it('denies access when user lacks required route permission', () => {
    expect(canAccessRoute(ROUTES.orgStructure, [PERMISSION.SELF_KEYS])).toBe(false)
  })

  it('allows unknown routes without explicit meta', () => {
    expect(canAccessRoute('/unknown', [PERMISSION.SELF_KEYS])).toBe(true)
  })
})

describe('getDefaultHomePath', () => {
  it('returns first matching home candidate path', () => {
    expect(getDefaultHomePath([PERMISSION.BUDGET_READ])).toBe(ROUTES.budgetOverview)
  })

  it('falls back to cost dashboard when no candidate matches', () => {
    expect(getDefaultHomePath([PERMISSION.API_CALL])).toBe(ROUTES.dashboardCost)
  })
})

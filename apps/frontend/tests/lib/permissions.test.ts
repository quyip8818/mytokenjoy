import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { ROUTES } from '@/config/routes'
import { PERMISSION } from '@/lib/permission-keys'
import {
  canAccessRoute,
  canWriteSession,
  getDefaultHomePath,
  hasPermission,
  isReadOnlySession,
} from '@/lib/permissions'

const manifestPath = join(
  import.meta.dirname,
  '../../../../packages/contracts/permission/manifest.json',
)

const manifest = JSON.parse(readFileSync(manifestPath, 'utf8')) as {
  capabilities: string[]
  platformCapabilities?: string[]
}

describe('manifest contract', () => {
  it('matches backend manifest capability count', () => {
    const allCaps = [...manifest.capabilities, ...(manifest.platformCapabilities ?? [])]
    const keys = Object.values(PERMISSION)
    expect(keys).toHaveLength(allCaps.length)
    expect(new Set(keys)).toEqual(new Set(allCaps))
  })
})

describe('hasPermission', () => {
  it.each([
    { user: [PERMISSION.ORG_ADMIN], required: PERMISSION.ORG_ADMIN, expected: true },
    {
      user: [PERMISSION.SELF_KEYS],
      required: [PERMISSION.ORG_ADMIN, PERMISSION.SELF_KEYS],
      expected: true,
    },
    { user: [PERMISSION.SELF_KEYS], required: PERMISSION.ORG_ADMIN, expected: false },
  ])('matches required permissions ($expected)', ({ user, required, expected }) => {
    expect(hasPermission(user, required)).toBe(expected)
  })
})

describe('isReadOnlySession', () => {
  it('returns false when server marks session writable', () => {
    expect(isReadOnlySession([PERMISSION.ORG_ADMIN], false)).toBe(false)
  })

  it('returns true when server marks session read-only', () => {
    expect(isReadOnlySession([PERMISSION.ORG_READ], true)).toBe(true)
  })

  it('returns false for wildcard permissions when writable', () => {
    expect(isReadOnlySession(['*'], false)).toBe(false)
  })
})

describe('canWriteSession', () => {
  it('mirrors server readOnly flag', () => {
    expect(canWriteSession([PERMISSION.ORG_ADMIN], false)).toBe(true)
    expect(canWriteSession([PERMISSION.ORG_READ], true)).toBe(false)
  })
})

describe('canAccessRoute', () => {
  it('allows access when user has required route permission', () => {
    expect(canAccessRoute(ROUTES.orgStructure, [PERMISSION.ORG_MANAGE])).toBe(true)
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
    expect(getDefaultHomePath([PERMISSION.BUDGET_READ])).toBe(ROUTES.budget)
  })

  it('falls back to permission-free route when no admin route matches', () => {
    expect(getDefaultHomePath([PERMISSION.API_CALL])).toBe(ROUTES.myKeys)
  })
})

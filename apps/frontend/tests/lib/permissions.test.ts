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
  '../../../backend/internal/infra/permission/manifest.json',
)

const manifest = JSON.parse(readFileSync(manifestPath, 'utf8')) as {
  capabilities: string[]
}

describe('manifest contract', () => {
  it('matches backend manifest capability count', () => {
    const keys = Object.values(PERMISSION)
    expect(keys).toHaveLength(manifest.capabilities.length)
    expect(new Set(keys)).toEqual(new Set(manifest.capabilities))
  })
})

describe('hasPermission', () => {
  it.each([
    { user: [PERMISSION.ORG_STRUCTURE], required: PERMISSION.ORG_STRUCTURE, expected: true },
    {
      user: [PERMISSION.SELF_KEYS],
      required: [PERMISSION.ORG_STRUCTURE, PERMISSION.SELF_KEYS],
      expected: true,
    },
    { user: [PERMISSION.SELF_KEYS], required: PERMISSION.ORG_STRUCTURE, expected: false },
  ])('matches required permissions ($expected)', ({ user, required, expected }) => {
    expect(hasPermission(user, required)).toBe(expected)
  })
})

describe('isReadOnlySession', () => {
  it('returns false when server marks session writable', () => {
    expect(isReadOnlySession([PERMISSION.ORG_STRUCTURE], false)).toBe(false)
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
    expect(canWriteSession([PERMISSION.ORG_STRUCTURE], false)).toBe(true)
    expect(canWriteSession([PERMISSION.ORG_READ], true)).toBe(false)
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

  it('returns null when no home candidate matches', () => {
    expect(getDefaultHomePath([PERMISSION.API_CALL])).toBeNull()
  })
})

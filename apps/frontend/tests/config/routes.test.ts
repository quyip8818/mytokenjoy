import { describe, expect, it } from 'vitest'
import {
  APP_ROUTES,
  NAV_GROUP_LAYOUT,
  ROUTE_DEFINITIONS,
  ROUTE_META,
  ROUTES,
  getRouteMeta,
  routePermissions,
} from '@/config/routes'
import { NAV_GROUPS, ROUTE_TITLES } from '@/config/nav'

describe('routes config', () => {
  it('covers every ROUTE_META path in ROUTE_TITLES', () => {
    for (const meta of ROUTE_META) {
      expect(ROUTE_TITLES[meta.path]).toBe(meta.label)
    }
  })

  it('exposes routePermissions consistent with getRouteMeta', () => {
    for (const meta of ROUTE_META) {
      expect(routePermissions(meta.path)).toEqual([...meta.requiredPermissions])
      expect(getRouteMeta(meta.path).label).toBe(meta.label)
    }
  })

  it('does not register home in APP_ROUTES', () => {
    expect(APP_ROUTES.some((entry) => entry.path === ROUTES.home)).toBe(false)
  })

  it('derives ROUTE_META and APP_ROUTES from ROUTE_DEFINITIONS', () => {
    expect(ROUTE_DEFINITIONS.length).toBe(ROUTE_META.length)
    expect(ROUTE_DEFINITIONS.length).toBe(APP_ROUTES.length)
  })

  it('covers every ROUTE_META path in navigation groups', () => {
    const navPaths = new Set(NAV_GROUP_LAYOUT.flatMap((group) => group.paths))
    for (const meta of ROUTE_META) {
      expect(navPaths.has(meta.path)).toBe(true)
    }
    expect(navPaths.size).toBe(ROUTE_META.length)
  })

  it('maps every ROUTE_DEFINITIONS key to ROUTES', () => {
    for (const definition of ROUTE_DEFINITIONS) {
      expect(ROUTES[definition.key as keyof typeof ROUTES]).toBe(definition.path)
    }
  })

  it('builds NAV_GROUPS with the same route count as ROUTE_META', () => {
    const itemCount = NAV_GROUPS.reduce((count, group) => count + group.items.length, 0)
    expect(itemCount).toBe(ROUTE_META.length)
  })
})

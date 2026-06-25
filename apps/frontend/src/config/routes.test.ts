import { describe, expect, it } from 'vitest'
import { APP_ROUTES, ROUTE_META, ROUTES, getRouteMeta, routePermissions } from '@/config/routes'
import { NAV_GROUP_LAYOUT, ROUTE_TITLES } from '@/config/nav'

describe('routes config', () => {
  it('keeps ROUTE_META aligned with APP_ROUTES paths', () => {
    const metaPaths = new Set(ROUTE_META.map((meta) => meta.path))
    const appPaths = APP_ROUTES.map((entry) => entry.path)

    expect(metaPaths.size).toBe(appPaths.length)
    for (const path of appPaths) {
      expect(metaPaths.has(path)).toBe(true)
    }
  })

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

  it('lists only registered paths in NAV_GROUP_LAYOUT', () => {
    const metaPaths = new Set(ROUTE_META.map((meta) => meta.path))
    for (const group of NAV_GROUP_LAYOUT) {
      for (const path of group.paths) {
        expect(metaPaths.has(path)).toBe(true)
      }
    }
  })

  it('includes every navigable route in NAV_GROUP_LAYOUT', () => {
    const navPaths = new Set(NAV_GROUP_LAYOUT.flatMap((group) => group.paths))
    for (const meta of ROUTE_META) {
      expect(navPaths.has(meta.path)).toBe(true)
    }
  })

  it('defines unique paths in ROUTE_META', () => {
    const paths = ROUTE_META.map((meta) => meta.path)
    expect(new Set(paths).size).toBe(paths.length)
  })

  it('does not register home in APP_ROUTES', () => {
    expect(APP_ROUTES.some((entry) => entry.path === ROUTES.home)).toBe(false)
  })
})

import { describe, expect, it } from 'vitest'
import { APP_ROUTES, ROUTE_META, ROUTES, getRouteMeta, routePermissions } from '@/config/routes'
import { ROUTE_TITLES } from '@/config/nav'

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
})

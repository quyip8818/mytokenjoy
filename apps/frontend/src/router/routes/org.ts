import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { authLayoutRoute } from '../auth-layout'

export const orgDataSourceRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/org/data-source',
  component: lazyRouteComponent(() => import('@/routes/org/data-source')),
})

export const orgStructureRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/org/structure',
  component: lazyRouteComponent(() => import('@/routes/org/structure')),
})

export const orgRolesRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/org/roles',
  component: lazyRouteComponent(() => import('@/routes/org/roles')),
})

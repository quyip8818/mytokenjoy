import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { authLayoutRoute } from '../auth-layout'

export const auditOperationsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/audit/operations',
  component: lazyRouteComponent(() => import('@/routes/audit/operations')),
})

export const auditCallsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/audit/calls',
  component: lazyRouteComponent(() => import('@/routes/audit/calls')),
})

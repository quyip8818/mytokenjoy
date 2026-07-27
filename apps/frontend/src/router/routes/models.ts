import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { authLayoutRoute } from '../auth-layout'

export const modelsListRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/models/list',
  component: lazyRouteComponent(() => import('@/routes/models/list')),
})

export const modelsRoutingRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/models/routing',
  component: lazyRouteComponent(() => import('@/routes/models/routing')),
})

import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { authLayoutRoute } from '../auth-layout'

export const budgetRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/budget',
  component: lazyRouteComponent(() => import('@/routes/budget')),
})

export const budgetAlertsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/budget/alerts',
  component: lazyRouteComponent(() => import('@/routes/budget/alerts')),
})

export const billingRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/billing',
  component: lazyRouteComponent(() => import('@/routes/billing')),
})

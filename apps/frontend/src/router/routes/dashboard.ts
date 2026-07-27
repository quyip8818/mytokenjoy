import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { z } from 'zod'
import { authLayoutRoute } from '../auth-layout'

export const dashboardCostRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/dashboard/cost',
  validateSearch: z.object({
    dept: z.string().optional(),
  }),
  component: lazyRouteComponent(() => import('@/routes/dashboard/cost')),
})

export const dashboardUsageRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/dashboard/usage',
  validateSearch: z.object({
    dept: z.string().optional(),
  }),
  component: lazyRouteComponent(() => import('@/routes/dashboard/usage')),
})

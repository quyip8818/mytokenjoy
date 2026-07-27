import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { z } from 'zod'
import { rootRoute } from '../root'

export const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: lazyRouteComponent(() => import('@/routes/auth/login')),
})

export const inviteAcceptRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/invite/accept',
  validateSearch: z.object({
    code: z.string().optional(),
  }),
  component: lazyRouteComponent(() => import('@/routes/auth/invite-accept')),
})

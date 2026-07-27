import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { z } from 'zod'
import { authLayoutRoute } from '../auth-layout'

export const keysPlatformRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/keys/platform',
  validateSearch: z.object({
    highlight: z.string().optional(),
    projectId: z.string().optional(),
  }),
  component: lazyRouteComponent(() => import('@/routes/keys/platform')),
})

export const approvalsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/approvals',
  component: lazyRouteComponent(() => import('@/routes/approval/index')),
})

export const keysProviderRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/keys/provider',
  component: lazyRouteComponent(() => import('@/routes/keys/provider')),
})

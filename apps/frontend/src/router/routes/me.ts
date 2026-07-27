import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { z } from 'zod'
import { authLayoutRoute } from '../auth-layout'

export const meKeysRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/me/keys',
  component: lazyRouteComponent(() => import('@/routes/me/keys')),
})

export const meUsageRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/me/usage',
  component: lazyRouteComponent(() => import('@/routes/me/usage')),
})

export const meSettingsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/me/settings',
  validateSearch: z.object({
    tab: z.enum(['account', 'security', 'notifications']).catch('account'),
  }),
  component: lazyRouteComponent(() => import('@/routes/me/settings')),
})

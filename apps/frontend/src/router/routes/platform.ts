import { createRoute, lazyRouteComponent } from '@tanstack/react-router'
import { authLayoutRoute } from '../auth-layout'

export const platformModelsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/platform/models',
  component: lazyRouteComponent(() => import('@/routes/platform/models')),
})

export const platformCompaniesRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/platform/companies',
  component: lazyRouteComponent(() => import('@/routes/platform/companies')),
})

export const platformCurrenciesRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/platform/currencies',
  component: lazyRouteComponent(() => import('@/routes/platform/currencies')),
})

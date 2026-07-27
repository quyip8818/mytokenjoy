import { createRouter } from '@tanstack/react-router'
import { RouteFallback } from '@/components/layout/route-fallback'
import { rootRoute } from './root'
import { authLayoutRoute } from './auth-layout'
import {
  loginRoute,
  indexRoute,
  dashboardRoute,
  suppliersRoute,
  supplierDetailRoute,
  modelsRoute,
  contractsRoute,
  ordersRoute,
  evaluationsRoute,
  newapiRoute,
  usersRoute,
  weightsRoute,
} from './routes'

const routeTree = rootRoute.addChildren([
  loginRoute,
  authLayoutRoute.addChildren([
    indexRoute,
    dashboardRoute,
    suppliersRoute,
    supplierDetailRoute,
    modelsRoute,
    contractsRoute,
    ordersRoute,
    evaluationsRoute,
    newapiRoute,
    usersRoute,
    weightsRoute,
  ]),
])

export const router = createRouter({
  routeTree,
  defaultPendingComponent: RouteFallback,
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

import { createRoute, lazyRouteComponent, redirect } from '@tanstack/react-router'
import { rootRoute } from './root'
import { authLayoutRoute } from './auth-layout'

// ─── Public ───

export const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: lazyRouteComponent(() => import('@/routes/auth/login')),
})

// ─── Index redirect ───

export const indexRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: '/dashboard' })
  },
})

// ─── Business routes ───

export const dashboardRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/dashboard',
  component: lazyRouteComponent(() => import('@/routes/dashboard/index')),
})

export const suppliersRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/suppliers',
  component: lazyRouteComponent(() => import('@/routes/suppliers/index')),
})

export const supplierDetailRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/suppliers/$id',
  component: lazyRouteComponent(() => import('@/routes/suppliers/detail')),
})

export const contractsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/contracts',
  component: lazyRouteComponent(() => import('@/routes/contracts/index')),
})

export const ordersRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/orders',
  component: lazyRouteComponent(() => import('@/routes/orders/index')),
})

export const evaluationsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/evaluations',
  component: lazyRouteComponent(() => import('@/routes/evaluations/index')),
})

export const usersRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/system/users',
  component: lazyRouteComponent(() => import('@/routes/system/users')),
})

export const weightsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/system/weights',
  component: lazyRouteComponent(() => import('@/routes/system/weights')),
})

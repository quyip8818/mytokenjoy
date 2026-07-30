import { createRouter } from '@tanstack/react-router'
import { RouteFallback } from '@/components/layout/route-fallback'
import { rootRoute } from './root'
import { authLayoutRoute } from './auth-layout'
import { loginRoute, inviteAcceptRoute } from './routes/auth'
import { setupRoute } from './routes/setup'
import { homeRoute } from './routes/home'
import { dashboardCostRoute, dashboardUsageRoute } from './routes/dashboard'
import { keysPlatformRoute, approvalsRoute, keysProviderRoute } from './routes/keys'
import { modelsListRoute, modelsRoutingRoute } from './routes/models'
import { budgetRoute, budgetAlertsRoute, billingRoute } from './routes/budget'
import { orgDataSourceRoute, orgStructureRoute, orgRolesRoute } from './routes/org'
import { auditOperationsRoute, auditCallsRoute } from './routes/audit'
import { meKeysRoute, meUsageRoute, meSettingsRoute } from './routes/me'
import { notificationsRoute } from './routes/notifications'
import {
  platformModelsRoute,
  platformCompaniesRoute,
  platformCurrenciesRoute,
} from './routes/platform'

const routeTree = rootRoute.addChildren([
  // Public routes
  loginRoute,
  inviteAcceptRoute,
  setupRoute,
  // Authenticated layout + its children
  authLayoutRoute.addChildren([
    homeRoute,
    dashboardCostRoute,
    dashboardUsageRoute,
    keysPlatformRoute,
    approvalsRoute,
    keysProviderRoute,
    modelsListRoute,
    modelsRoutingRoute,
    budgetRoute,
    budgetAlertsRoute,
    billingRoute,
    orgDataSourceRoute,
    orgStructureRoute,
    orgRolesRoute,
    auditOperationsRoute,
    auditCallsRoute,
    meKeysRoute,
    meUsageRoute,
    meSettingsRoute,
    notificationsRoute,
    platformModelsRoute,
    platformCompaniesRoute,
    platformCurrenciesRoute,
  ]),
])

export const router = createRouter({
  routeTree,
  defaultPendingComponent: RouteFallback,
  basepath: import.meta.env.BASE_URL,
})

// Register router for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

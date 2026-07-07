import { lazy, Suspense } from 'react'
import { BrowserRouter, Navigate, Routes, Route } from 'react-router'
import { LOGIN_PATH } from '@/config/auth'
import { MEMBER_ROUTE_DEFINITIONS, toMemberRouterPath } from '@/config/member-routes'
import { AppErrorBoundary } from '@/components/layout/app-error-boundary'
import { AppProviders } from '@/components/layout/app-providers'
import { HomeRedirect } from '@/components/layout/home-redirect'
import { AdminLayout } from '@/components/layout/admin-layout'
import { MemberLayout } from '@/components/layout/member-layout'
import { RouteFallback } from '@/components/layout/route-fallback'
import { SessionGate } from '@/features/session/session-gate'
import { APP_ROUTES, toRouterPath } from '@/config/routes'

const LoginPage = lazy(() => import('@/routes/auth/login'))

const lazyPages = APP_ROUTES.map((entry) => ({
  path: toRouterPath(entry.path),
  Page: lazy(entry.lazy),
}))

const memberLazyPages = MEMBER_ROUTE_DEFINITIONS.map((entry) => ({
  path: toMemberRouterPath(entry.path),
  Page: lazy(entry.lazy),
}))

function AuthenticatedRoutes() {
  return (
    <SessionGate>
      <Routes>
        <Route element={<AdminLayout />}>
          <Route index element={<HomeRedirect />} />
          {lazyPages.map(({ path, Page }) => (
            <Route key={path} path={path} element={<Page />} />
          ))}
          <Route path="billing" element={<Navigate to="/wallet" replace />} />
        </Route>
        <Route element={<MemberLayout />}>
          {memberLazyPages.map(({ path, Page }) => (
            <Route key={path} path={path} element={<Page />} />
          ))}
        </Route>
      </Routes>
    </SessionGate>
  )
}

export default function App() {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <AppErrorBoundary>
        <AppProviders>
          <Suspense fallback={<RouteFallback />}>
            <Routes>
              <Route
                path={LOGIN_PATH.slice(1)}
                element={
                  <Suspense fallback={<RouteFallback />}>
                    <LoginPage />
                  </Suspense>
                }
              />
              <Route path="*" element={<AuthenticatedRoutes />} />
            </Routes>
          </Suspense>
        </AppProviders>
      </AppErrorBoundary>
    </BrowserRouter>
  )
}

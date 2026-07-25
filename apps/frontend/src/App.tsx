import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import { LOGIN_PATH } from '@/config/auth'
import { APP_ROUTES, toRouterPath } from '@/config/routes'
import { AppErrorBoundary } from '@/components/layout/app-error-boundary'
import { AppProviders } from '@/components/layout/app-providers'
import { HomeRedirect } from '@/components/layout/home-redirect'
import { AdminLayout } from '@/components/layout/admin-layout'
import { RouteFallback } from '@/components/layout/route-fallback'
import { SessionGate } from '@/features/session'

const LoginPage = lazy(() => import('@/routes/auth/login'))
const InviteAcceptPage = lazy(() => import('@/routes/auth/invite-accept'))
const AuthEmbedPage = lazy(() => import('@/routes/auth/embed'))

const lazyPages = APP_ROUTES.map((entry) => ({
  path: toRouterPath(entry.path),
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
              <Route
                path="invite/accept"
                element={
                  <Suspense fallback={<RouteFallback />}>
                    <InviteAcceptPage />
                  </Suspense>
                }
              />
              <Route
                path="auth/embed"
                element={
                  <Suspense fallback={<RouteFallback />}>
                    <AuthEmbedPage />
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

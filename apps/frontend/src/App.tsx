import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import { LOGIN_PATH } from '@/config/auth'
import { AppErrorBoundary } from '@/components/layout/app-error-boundary'
import { HomeRedirect } from '@/components/layout/home-redirect'
import { AdminLayout } from '@/components/layout/admin-layout'
import { RouteFallback } from '@/components/layout/route-fallback'
import { APP_ROUTES, toRouterPath } from '@/config/routes'

const LoginPage = lazy(() => import('@/routes/auth/login'))

const lazyPages = APP_ROUTES.map((entry) => ({
  path: toRouterPath(entry.path),
  Page: lazy(entry.lazy),
}))

export default function App() {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <AppErrorBoundary>
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
            <Route element={<AdminLayout />}>
              <Route index element={<HomeRedirect />} />
              {lazyPages.map(({ path, Page }) => (
                <Route key={path} path={path} element={<Page />} />
              ))}
            </Route>
          </Routes>
        </Suspense>
      </AppErrorBoundary>
    </BrowserRouter>
  )
}

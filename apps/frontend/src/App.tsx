import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router'
import { HomeRedirect } from '@/components/layout/home-redirect'
import { AdminLayout } from '@/components/layout/admin-layout'
import { RouteFallback } from '@/components/layout/route-fallback'
import { APP_ROUTES, toRouterPath } from '@/config/routes'

const lazyPages = APP_ROUTES.map((entry) => ({
  path: toRouterPath(entry.path),
  Page: lazy(entry.lazy),
}))

export default function App() {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <Suspense fallback={<RouteFallback />}>
        <Routes>
          <Route element={<AdminLayout />}>
            <Route index element={<HomeRedirect />} />
            {lazyPages.map(({ path, Page }) => (
              <Route key={path} path={path} element={<Page />} />
            ))}
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

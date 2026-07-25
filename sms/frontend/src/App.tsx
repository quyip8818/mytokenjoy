import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import { AppProviders } from '@/components/layout/app-providers'
import { AdminLayout } from '@/components/layout/admin-layout'
import { RouteFallback } from '@/components/layout/route-fallback'
import { useSession } from '@/features/session'
import { ROUTE_DEFINITIONS } from '@/config/routes'
import LoginPage from '@/routes/auth/login'

const lazyPages = ROUTE_DEFINITIONS.map((entry) => ({
  path: entry.path.slice(1),
  Page: lazy(entry.lazy),
}))

function RequireAuth({ children }: { children: React.ReactNode }) {
  const user = useSession((s) => s.user)
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <AppProviders>
        <Suspense fallback={<RouteFallback />}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route
              path="/*"
              element={
                <RequireAuth>
                  <Routes>
                    <Route path="/" element={<Navigate to="/dashboard" replace />} />
                    <Route element={<AdminLayout />}>
                      {lazyPages.map(({ path, Page }) => (
                        <Route key={path} path={path} element={<Page />} />
                      ))}
                    </Route>
                  </Routes>
                </RequireAuth>
              }
            />
          </Routes>
        </Suspense>
      </AppProviders>
    </BrowserRouter>
  )
}

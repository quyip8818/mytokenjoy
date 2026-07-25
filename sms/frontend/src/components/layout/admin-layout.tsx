import { Outlet } from 'react-router'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { RouteErrorBoundary } from './error-boundary'

export function AdminLayout() {
  return (
    <div className="flex h-screen">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-auto p-6">
          <RouteErrorBoundary>
            <Outlet />
          </RouteErrorBoundary>
        </main>
      </div>
    </div>
  )
}

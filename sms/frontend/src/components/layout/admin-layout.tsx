import { Outlet } from 'react-router'
import { Toaster } from 'sonner'
import { Sidebar } from './sidebar'
import { SidebarLayoutProvider } from './sidebar-layout-provider'
import { Header } from './header'

export function AdminLayout() {
  return (
    <SidebarLayoutProvider>
      <div className="flex h-screen bg-background">
        <Sidebar />
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
          <Header />
          <main className="flex min-h-0 flex-1 flex-col overflow-hidden p-8">
            <div className="min-h-0 flex-1 overflow-auto">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
      <Toaster theme="light" position="top-center" richColors />
    </SidebarLayoutProvider>
  )
}

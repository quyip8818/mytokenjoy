/* eslint-disable react-refresh/only-export-components */
import { Outlet, createRoute, redirect } from '@tanstack/react-router'
import { Toaster } from 'sonner'
import { useSession } from '@/features/session'
import { Sidebar } from '@/components/layout/sidebar'
import { SidebarLayoutProvider } from '@/components/layout/sidebar-layout-provider'
import { Header } from '@/components/layout/header'
import { WorkflowProvider, WorkflowPanelStack } from '@/features/workflow'
import { rootRoute } from './root'

function AuthenticatedLayout() {
  return (
    <WorkflowProvider>
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
        <WorkflowPanelStack />
        <Toaster theme="light" position="top-center" richColors />
      </SidebarLayoutProvider>
    </WorkflowProvider>
  )
}

export const authLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  beforeLoad: () => {
    const user = useSession.getState().user
    if (!user) throw redirect({ to: '/login' })
  },
  component: AuthenticatedLayout,
})

import { Outlet } from 'react-router'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider } from '@/features/query'
import { AuthSessionProvider, SessionNavigationBridge } from '@/features/session'
import { AuthUnauthorizedBridge } from '@/components/auth/auth-unauthorized-bridge'
import { WorkflowProvider } from '@/features/workflow/workflow-context'
import { WorkflowPanelStack } from '@/features/workflow/components/workflow-panel-stack'
import { Toaster } from '@/components/ui/sonner'
import { Sidebar } from './sidebar'
import { SidebarLayoutProvider } from './sidebar-layout-provider'
import { Header } from './header'

function AdminShell() {
  return (
    <SidebarLayoutProvider>
      <WorkflowProvider>
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
        <Toaster theme="light" />
      </WorkflowProvider>
    </SidebarLayoutProvider>
  )
}

export function AdminLayout() {
  return (
    <ApiProvider apis={defaultApis}>
      <QueryProvider>
        <AuthSessionProvider>
          <AuthUnauthorizedBridge />
          <SessionNavigationBridge />
          <AdminShell />
        </AuthSessionProvider>
      </QueryProvider>
    </ApiProvider>
  )
}

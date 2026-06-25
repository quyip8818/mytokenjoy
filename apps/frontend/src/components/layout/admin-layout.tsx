import { Outlet } from 'react-router'
import { USE_MOCKS } from '@/config/app'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { AuthSessionProvider } from '@/features/session'
import { WorkflowProvider } from '@/features/workflow/workflow-context'
import { WorkflowPanelStack } from '@/features/workflow/components/workflow-panel-stack'
import { Toaster } from '@/components/ui/sonner'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { LazyDemoShellBoundary } from './lazy-demo-shell'

function AdminShell() {
  return (
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
  )
}

export function AdminLayout() {
  return (
    <ApiProvider apis={defaultApis}>
      {USE_MOCKS ? (
        <LazyDemoShellBoundary>
          <AdminShell />
        </LazyDemoShellBoundary>
      ) : (
        <AuthSessionProvider>
          <AdminShell />
        </AuthSessionProvider>
      )}
    </ApiProvider>
  )
}

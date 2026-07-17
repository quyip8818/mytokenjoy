import { Outlet } from 'react-router'
import { AppErrorBoundary } from '@/components/layout/app-error-boundary'
import { WorkflowProvider } from '@/features/workflow'
import { WorkflowPanelStack } from '@/features/workflow'
import { Toaster } from '@/components/ui/sonner'
import { useSession } from '@/features/session'
import { TrialBanner } from '@/features/trial'
import { Sidebar } from './sidebar'
import { SidebarLayoutProvider } from './sidebar-layout-provider'
import { Header } from './header'

export function AdminLayout() {
  const { companyType } = useSession()

  return (
    <SidebarLayoutProvider>
      <WorkflowProvider>
        <div className="flex h-screen bg-background">
          <Sidebar />
          <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
            {companyType === 'trial' && <TrialBanner />}
            <Header />
            <main className="flex min-h-0 flex-1 flex-col overflow-hidden p-8">
              <div className="min-h-0 flex-1 overflow-auto">
                <AppErrorBoundary>
                  <Outlet />
                </AppErrorBoundary>
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

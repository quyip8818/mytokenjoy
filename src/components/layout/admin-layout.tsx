import { Outlet } from 'react-router'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { WorkflowProvider } from '@/features/workflow/workflow-context'
import { WorkflowPanelStack } from '@/features/workflow/components/workflow-panel-stack'
import { DemoProvider } from '@/features/demo'
import { PageContextProvider } from '@/features/layout/page-context-provider'
import { Toaster } from '@/components/ui/sonner'

export function AdminLayout() {
  return (
    <DemoProvider>
      <PageContextProvider>
        <WorkflowProvider>
          <div className="flex h-screen bg-background">
            <Sidebar />
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
              <Header />
              <main className="relative flex min-h-0 flex-1 flex-col overflow-hidden p-8">
                <div className="min-h-0 flex-1 overflow-auto">
                  <Outlet />
                </div>
                <WorkflowPanelStack />
              </main>
            </div>
          </div>
          <Toaster theme="light" />
        </WorkflowProvider>
      </PageContextProvider>
    </DemoProvider>
  )
}

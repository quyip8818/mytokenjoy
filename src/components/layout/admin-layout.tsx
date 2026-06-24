import { Outlet } from 'react-router'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { WorkflowProvider } from '@/features/workflow/workflow-context'
import { WorkflowPanelStack } from '@/features/workflow/components/workflow-panel-stack'
import { DemoProvider } from '@/features/demo'
import { Toaster } from '@/components/ui/sonner'

export function AdminLayout() {
  return (
    <DemoProvider>
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
    </DemoProvider>
  )
}

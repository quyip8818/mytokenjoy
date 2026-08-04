/* eslint-disable react-refresh/only-export-components */
import { useEffect } from 'react'
import { Outlet, createRoute, useRouter, useRouterState } from '@tanstack/react-router'
import { WorkflowProvider, WorkflowPanelStack } from '@/features/workflow'
import { Toaster } from '@/components/ui/sonner'
import { useSession } from '@/features/session'
import { TrialBanner } from '@/components/layout/trial-banner'
import { AnnouncementDialog } from '@/features/announcements'
import { Sidebar } from '@/components/layout/sidebar'
import { SidebarLayoutProvider } from '@/components/layout/sidebar-layout-provider'
import { Header } from '@/components/layout/header'
import { ApiError } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'
import { canAccessRoute, getDefaultHomePath } from '@/lib/permissions'
import { ErrorState } from '@/components/ui/error-state'
import { rootRoute } from './root'

/**
 * Authenticated layout route.
 * ponytail: session 检查留在组件级——因为 session 来自 React Context (TanStack Query)，
 * 不能在 beforeLoad 里同步读取。升级路径：session 改 zustand 后可以移到 beforeLoad。
 */
function AuthenticatedLayout() {
  const { companyType, sessionError, loading, permissions, refreshSession } = useSession()
  const router = useRouter()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  const isUnauthorized = sessionError instanceof ApiError && sessionError.status === 401

  // Redirect to login on 401
  useEffect(() => {
    if (isUnauthorized) {
      window.location.replace(LOGIN_PATH)
    }
  }, [isUnauthorized])

  // Permission watcher: redirect when permissions change and current route is no longer accessible
  useEffect(() => {
    if (loading || !permissions.length) return
    if (!canAccessRoute(pathname, permissions)) {
      const home = getDefaultHomePath(permissions)
      void router.navigate({ to: home ?? '/', replace: true })
    }
  }, [permissions, pathname, loading, router])

  if (loading) return null
  if (isUnauthorized) return null

  if (sessionError) {
    return (
      <div className="flex min-h-screen items-center justify-center p-8">
        <ErrorState message={sessionError.message} onRetry={() => void refreshSession()} />
      </div>
    )
  }

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
                <Outlet />
              </div>
            </main>
          </div>
        </div>
        <WorkflowPanelStack />
        <Toaster theme="light" />
        <AnnouncementDialog />
      </WorkflowProvider>
    </SidebarLayoutProvider>
  )
}

export const authLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  component: AuthenticatedLayout,
})

import { Outlet, NavLink, Link } from 'react-router'
import { useSession } from '@/features/session'
import { WorkflowProvider } from '@/features/workflow'
import { WorkflowPanelStack } from '@/features/workflow'
import { MEMBER_ROUTE_DEFINITIONS } from '@/config/routes'
import { cn } from '@/lib/utils'
import { Toaster } from '@/components/ui/sonner'
import { AppErrorBoundary } from '@/components/layout/app-error-boundary'

export function MemberLayout() {
  const { member } = useSession()
  const displayName = member?.name ?? '—'
  const initial = displayName.slice(0, 1)

  return (
    <WorkflowProvider>
      <div className="flex h-dvh bg-background">
        <aside className="flex w-48 shrink-0 flex-col border-r border-border bg-card">
          <div className="px-5 pt-5 pb-4">
            <Link to="/" className="group block">
              <img src="/logo.png" alt="TokenJoy" className="h-7 w-auto transition-transform group-hover:scale-105" />
              <p className="mt-1 text-[10px] text-muted-foreground">个人中心</p>
            </Link>
          </div>
          <nav className="flex-1 space-y-0.5 px-3">
            {MEMBER_ROUTE_DEFINITIONS.map((item) => {
              const Icon = item.icon
              return (
                <NavLink
                  key={item.path}
                  to={item.path}
                  end={item.navEnd}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-2.5 rounded-md px-3 py-2 text-sm',
                      isActive
                        ? 'bg-muted font-medium text-foreground'
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                    )
                  }
                >
                  <Icon className="size-4 shrink-0" strokeWidth={1.5} />
                  {item.label}
                </NavLink>
              )
            })}
          </nav>
          <div className="border-t border-border p-4">
            <div className="flex items-center gap-2">
              <div className="flex size-7 items-center justify-center rounded-md bg-primary text-[10px] font-medium text-primary-foreground">
                {initial}
              </div>
              <div>
                <p className="text-sm font-medium">{displayName}</p>
                <p className="text-xs text-muted-foreground">{member?.departmentName ?? '—'}</p>
              </div>
            </div>
          </div>
        </aside>
        <div className="flex flex-1 flex-col overflow-hidden">
          <main className="flex-1 overflow-auto p-8">
            <AppErrorBoundary>
              <Outlet />
            </AppErrorBoundary>
          </main>
        </div>
      </div>
      <WorkflowPanelStack />
      <Toaster theme="light" />
    </WorkflowProvider>
  )
}

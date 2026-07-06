import { Outlet, NavLink } from 'react-router'
import { useSession } from '@/features/session'
import { MEMBER_ROUTE_DEFINITIONS } from '@/config/member-routes'
import { cn } from '@/lib/utils'

export function MemberLayout() {
  const { member } = useSession()
  const displayName = member?.name ?? '—'
  const initial = displayName.slice(0, 1)

  return (
    <div className="flex h-dvh bg-background">
      <aside className="flex w-48 shrink-0 flex-col border-r border-border bg-card">
        <div className="px-5 pt-5 pb-4">
          <h1 className="text-base font-semibold text-primary">TokenJoy</h1>
          <p className="mt-0.5 text-xs text-muted-foreground">个人中心</p>
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
          <Outlet />
        </main>
      </div>
    </div>
  )
}

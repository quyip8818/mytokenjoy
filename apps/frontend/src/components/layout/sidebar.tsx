import { NavLink, useLocation } from 'react-router'
import { cn } from '@/lib/utils'
import { getVisibleNavGroups } from '@/config/nav'
import { useApprovalPendingCount } from '@/hooks/use-approval-pending-count'
import { usePermissions } from '@/hooks/use-permissions'

export function Sidebar() {
  const location = useLocation()
  const { permissions } = usePermissions()
  const navGroups = getVisibleNavGroups(permissions)
  const approvalPendingCount = useApprovalPendingCount()

  const getBadge = (badgeKey?: string) => {
    if (badgeKey === 'approvalPending' && approvalPendingCount > 0) {
      return approvalPendingCount
    }
    return 0
  }

  return (
    <aside
      className="relative flex w-60 flex-col overflow-hidden border-r border-sidebar-border bg-sidebar"
      style={{ boxShadow: 'var(--shadow-sidebar)' }}
    >
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-primary/[0.02] via-transparent to-sky-500/[0.015]" />
      <div className="pointer-events-none absolute -bottom-24 -left-24 h-64 w-64 rounded-full bg-primary/4 blur-3xl" />
      <div className="pointer-events-none absolute -top-12 -right-12 h-40 w-40 rounded-full bg-sky-400/4 blur-3xl" />

      <div className="relative z-10 px-5 pt-6 pb-4">
        <h1 className="text-xl font-extrabold tracking-tight text-gradient">TokenJoy</h1>
        <p className="mt-0.5 text-[11px] text-muted-foreground">LLM API 管理平台</p>
      </div>

      <nav className="relative z-10 flex-1 space-y-5 overflow-y-auto px-3 pb-4">
        {navGroups.map((group) => {
          const isGroupActive = group.items.some((item) => location.pathname.startsWith(item.path))
          return (
            <div key={group.group}>
              <div
                className={cn(
                  'mb-1.5 px-2 text-[10px] font-semibold uppercase tracking-wider',
                  isGroupActive ? 'text-primary' : 'text-muted-foreground',
                  group.collapsed && 'text-muted-foreground/60',
                )}
              >
                {group.group}
              </div>
              <div className="space-y-0.5">
                {group.items.map((item) => {
                  const Icon = item.icon
                  const badge = getBadge(item.badgeKey)
                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      className={({ isActive }) =>
                        cn(
                          'flex items-center gap-2.5 rounded-lg px-3 py-2 text-[13px] transition-all duration-150',
                          isActive
                            ? 'bg-sidebar-accent font-medium text-sidebar-accent-foreground ring-1 ring-primary/5'
                            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                        )
                      }
                    >
                      <Icon className="h-4 w-4 shrink-0" />
                      <span className="flex-1">{item.label}</span>
                      {badge > 0 && (
                        <span className="inline-flex min-w-[18px] items-center justify-center rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold text-primary-foreground">
                          {badge}
                        </span>
                      )}
                    </NavLink>
                  )
                })}
              </div>
            </div>
          )
        })}
      </nav>
    </aside>
  )
}

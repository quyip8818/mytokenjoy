import { Link, useLocation } from 'react-router'
import { cn } from '@/lib/utils'
import { NAV_GROUP_LAYOUT } from '@/config/routes'
import { useSession } from '@/features/session'

export function Sidebar() {
  const location = useLocation()
  const role = useSession((s) => s.user?.role)

  return (
    <aside className="flex w-56 flex-col border-r bg-white">
      <div className="flex h-14 items-center border-b px-4">
        <h1 className="text-lg font-semibold">SMS</h1>
      </div>
      <nav className="flex-1 overflow-auto py-4">
        {NAV_GROUP_LAYOUT.map((group) => {
          const visibleItems = group.items.filter(
            (item) => !item.requiredRoles || (role && item.requiredRoles.includes(role)),
          )
          if (visibleItems.length === 0) return null
          return (
            <div key={group.group} className="mb-4">
              <div className="px-4 pb-1 text-xs font-medium text-muted-foreground">
                {group.group}
              </div>
              {visibleItems.map((item) => {
                const active = location.pathname === item.path
                const Icon = item.icon
                return (
                  <Link
                    key={item.key}
                    to={item.path}
                    className={cn(
                      'mx-2 flex items-center gap-2 rounded-md px-3 py-2 text-sm',
                      active
                        ? 'bg-primary/10 font-medium text-primary'
                        : 'text-muted-foreground hover:bg-muted',
                    )}
                  >
                    <Icon className="h-4 w-4" />
                    {item.label}
                  </Link>
                )
              })}
            </div>
          )
        })}
      </nav>
    </aside>
  )
}

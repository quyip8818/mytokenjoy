import { Outlet, NavLink } from 'react-router'
import { cn } from '@/lib/utils'
import { LayoutDashboard, Key, FileText } from 'lucide-react'

const navItems = [
  { label: '工作台', path: '/me', icon: LayoutDashboard },
  { label: '我的 Key', path: '/me/keys', icon: Key },
  { label: '使用记录', path: '/me/call-logs', icon: FileText },
]

export function MemberLayout() {
  return (
    <div className="flex h-dvh bg-background">
      {/* Sidebar */}
      <aside className="flex w-48 shrink-0 flex-col border-r border-border bg-card">
        <div className="px-5 pt-5 pb-4">
          <h1 className="text-base font-semibold text-primary">TokenJoy</h1>
          <p className="mt-0.5 text-xs text-muted-foreground">个人中心</p>
        </div>
        <nav className="flex-1 px-3 space-y-0.5">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.path}
                to={item.path}
                end
                className={({ isActive }) =>
                  cn(
                    'flex items-center gap-2.5 rounded-md px-3 py-2 text-sm',
                    isActive
                      ? 'bg-muted text-foreground font-medium'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted',
                  )
                }
              >
                <Icon className="size-4 shrink-0" strokeWidth={1.5} />
                {item.label}
              </NavLink>
            )
          })}
        </nav>
        {/* User info at bottom */}
        <div className="border-t border-border p-4">
          <div className="flex items-center gap-2">
            <div className="flex size-7 items-center justify-center rounded-md bg-primary text-[10px] font-medium text-primary-foreground">
              张
            </div>
            <div>
              <p className="text-sm font-medium">张三</p>
              <p className="text-xs text-muted-foreground">后端组</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main */}
      <div className="flex flex-1 flex-col overflow-hidden">
        <main className="flex-1 overflow-auto p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

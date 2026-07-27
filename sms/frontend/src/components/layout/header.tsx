import { LogOut } from 'lucide-react'
import { useLocation } from 'react-router'
import { ROUTE_TITLES } from '@/config/routes'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'

export function Header() {
  const { user, logout } = useSession()
  const { authApi } = useApis()
  const location = useLocation()
  const title = ROUTE_TITLES[location.pathname] || '控制台'

  const handleLogout = async () => {
    try {
      await authApi.logout()
    } catch {
      /* ignore */
    }
    logout()
  }

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-card px-8">
      <h1 className="truncate text-sm font-medium text-foreground">{title}</h1>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5">
          <span className="text-sm text-foreground">{user?.realName}</span>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <LogOut className="size-3.5" />
          退出
        </button>
      </div>
    </header>
  )
}

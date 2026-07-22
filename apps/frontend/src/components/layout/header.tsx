import { useCallback } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { User } from 'lucide-react'
import { ROUTE_TITLES } from '@/config/nav'
import { useSession } from '@/features/session'
import { HeaderDevBackendToolbar } from './header-dev-backend-chrome'
import { NotificationInbox } from './notification-inbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { authApi } from '@/api/auth'

/** Company badge — read-only display of current company context. */
function HeaderCompanyChip() {
  const { member } = useSession()
  const companyName = member?.name ?? '管理员'
  const initial = companyName.charAt(0) || '管'

  return (
    <div className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5">
      <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary text-[10px] font-medium text-primary-foreground">
        {initial}
      </div>
      <span className="text-sm text-foreground">{companyName}</span>
    </div>
  )
}

/** User menu — account settings & logout. */
function HeaderUserMenu() {
  const navigate = useNavigate()

  const handleLogout = useCallback(async () => {
    await authApi.logout()
    navigate('/login', { replace: true })
  }, [navigate])

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border transition-colors hover:bg-muted"
          aria-label="用户菜单"
        >
          <User className="h-4 w-4 text-muted-foreground" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onSelect={() => navigate('/me/account')}>账户设置</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={handleLogout}>退出登录</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function Header() {
  const location = useLocation()
  const title = ROUTE_TITLES[location.pathname] || '控制台'

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-card px-8">
      <h1 className="truncate text-sm font-medium text-foreground">{title}</h1>
      <div className="flex items-center gap-3">
        <NotificationInbox />
        <HeaderCompanyChip />
        <HeaderUserMenu />
        <HeaderDevBackendToolbar />
      </div>
    </header>
  )
}

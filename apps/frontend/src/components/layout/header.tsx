import { useCallback } from 'react'
import { useLocation, useNavigate } from 'react-router'
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

function HeaderUserChip() {
  const { member } = useSession()
  const navigate = useNavigate()
  const displayName = member?.name ?? '管理员'
  const initial = displayName.charAt(0) || '管'

  const handleLogout = useCallback(async () => {
    await authApi.logout()
    navigate('/login', { replace: true })
  }, [navigate])

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 transition-colors hover:bg-muted"
        >
          <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary text-[10px] font-medium text-primary-foreground">
            {initial}
          </div>
          <span className="text-sm text-foreground">{displayName}</span>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onSelect={() => navigate('/account')}>账户设置</DropdownMenuItem>
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
        <HeaderUserChip />
        <HeaderDevBackendToolbar />
      </div>
    </header>
  )
}

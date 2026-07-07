import { useLocation } from 'react-router'
import { ROUTE_TITLES } from '@/config/nav'
import { useSession } from '@/features/session/use-session'
import { HeaderDevBackendToolbar } from './header-dev-backend-chrome'

function HeaderUserChip() {
  const { member } = useSession()
  const displayName = member?.name ?? '管理员'
  const initial = displayName.charAt(0) || '管'

  return (
    <div className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 transition-colors hover:bg-muted">
      <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary text-[10px] font-medium text-primary-foreground">
        {initial}
      </div>
      <span className="text-sm text-foreground">{displayName}</span>
    </div>
  )
}

export function Header() {
  const location = useLocation()
  const title = ROUTE_TITLES[location.pathname] || '控制台'

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-card px-8">
      <h1 className="truncate text-sm font-medium text-foreground">{title}</h1>
      <div className="flex items-center gap-3">
        <HeaderUserChip />
        <HeaderDevBackendToolbar />
      </div>
    </header>
  )
}

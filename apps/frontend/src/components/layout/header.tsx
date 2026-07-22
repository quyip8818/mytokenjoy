import { useLocation, useNavigate } from 'react-router'
import { User } from 'lucide-react'
import { ROUTE_TITLES } from '@/config/nav'
import { useSession } from '@/features/session'
import { HeaderDevBackendToolbar } from './header-dev-backend-chrome'
import { NotificationInbox } from './notification-inbox'

/** Company badge — read-only display of current company context. */
function HeaderCompanyChip() {
  const { member, userName } = useSession()
  const companyName = member?.alias || userName || '管理员'
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

/** User tag — avatar + name, navigates to account page. */
function HeaderUserChip() {
  const navigate = useNavigate()
  const { member, userName } = useSession()
  const displayName = member?.alias || userName || '用户'
  const avatar = member?.avatar

  return (
    <button
      type="button"
      className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 transition-colors hover:bg-muted"
      aria-label="账户设置"
      onClick={() => navigate('/me/account')}
    >
      {avatar ? (
        <img src={avatar} alt="" className="h-6 w-6 rounded-md object-cover" />
      ) : (
        <div className="flex h-6 w-6 items-center justify-center rounded-md bg-muted">
          <User className="h-3.5 w-3.5 text-muted-foreground" />
        </div>
      )}
      <span className="text-sm text-foreground">{displayName}</span>
    </button>
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
        <HeaderUserChip />
        <HeaderDevBackendToolbar />
      </div>
    </header>
  )
}

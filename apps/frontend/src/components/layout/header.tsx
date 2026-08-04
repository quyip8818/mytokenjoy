import { useNavigate, useRouterState } from '@tanstack/react-router'
import { ROUTE_TITLES } from '@/config/nav'
import { useSession } from '@/features/session'
import { UserAvatar } from '@/components/ui/user-avatar'
import { HeaderDevBackendToolbar } from './header-dev-backend-chrome'
import { NotificationInbox } from './notification-inbox'

/** Reusable chip — avatar/icon + label, optionally clickable. */
function HeaderChip({
  avatar,
  label,
  onClick,
}: {
  avatar: React.ReactNode
  label: string
  onClick?: () => void
}) {
  const Comp = onClick ? 'button' : 'div'
  return (
    <Comp
      type={onClick ? 'button' : undefined}
      className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 shadow transition-colors hover:bg-muted"
      aria-label={onClick ? label : undefined}
      onClick={onClick}
    >
      {avatar}
      <span className="text-sm text-foreground">{label}</span>
    </Comp>
  )
}

/** Company badge — read-only display of current company context. */
function HeaderCompanyChip() {
  const { companyName } = useSession()
  const displayName = companyName || '管理员'
  const initial = displayName.charAt(0) || '管'

  return (
    <HeaderChip
      label={displayName}
      avatar={
        <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary text-[10px] font-medium text-primary-foreground">
          {initial}
        </div>
      }
    />
  )
}

/** User tag — avatar + name, navigates to account page. */
function HeaderUserChip() {
  const navigate = useNavigate()
  const { member, userName } = useSession()
  const displayName = member?.alias || userName || '用户'

  return (
    <HeaderChip
      label={displayName}
      avatar={<UserAvatar avatar={member?.avatar} fallback={displayName} size={24} />}
      onClick={() => navigate({ to: '/me/settings', search: { tab: 'account' } })}
    />
  )
}

export function Header() {
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const title = ROUTE_TITLES[pathname] || '控制台'

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

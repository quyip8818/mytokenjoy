import { useLocation } from 'react-router'
import { ROUTE_TITLES } from '@/config/nav'
import { usePageSubtitle } from '@/hooks/use-page-subtitle'
import { HeaderDemoBadge, HeaderDemoToolbar } from './header-demo-chrome'

export function Header() {
  const location = useLocation()
  const title = ROUTE_TITLES[location.pathname] || '控制台'
  const { subtitle } = usePageSubtitle()

  return (
    <header className="shrink-0 border-b border-border/60 bg-card/80 backdrop-blur-sm">
      <div className="flex h-14 items-center justify-between gap-4 px-8">
        <div className="flex min-w-0 items-center gap-3">
          <h1 className="truncate text-lg font-bold tracking-tight text-foreground">{title}</h1>
          <HeaderDemoBadge />
          {subtitle && (
            <span className="hidden truncate text-sm text-muted-foreground sm:inline">
              / {subtitle}
            </span>
          )}
        </div>
        <HeaderDemoToolbar />
      </div>
    </header>
  )
}

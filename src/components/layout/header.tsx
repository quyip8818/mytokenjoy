import { useLocation } from 'react-router'
import { ROUTE_TITLES } from '@/config/nav'
import { usePageContext } from '@/features/layout/use-page-context'
import { DemoBanner, DemoToolbar } from '@/features/demo'

export function Header() {
  const location = useLocation()
  const title = ROUTE_TITLES[location.pathname] || '控制台'
  const { subtitle } = usePageContext()

  return (
    <header className="shrink-0 border-b border-border/60 bg-card/80 backdrop-blur-sm">
      <DemoBanner />
      <div className="flex h-16 items-center justify-between px-8">
        <div>
          <h1 className="text-lg font-bold tracking-tight text-foreground">{title}</h1>
          {subtitle && <p className="text-sm text-muted-foreground mt-0.5">{subtitle}</p>}
        </div>
        <div className="flex items-center gap-3">
          <DemoToolbar />
        </div>
      </div>
    </header>
  )
}

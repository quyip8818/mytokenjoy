import type { ReactNode } from 'react'

interface DashboardPageLayoutProps {
  sidebar: ReactNode
  children: ReactNode
}

export function DashboardPageLayout({ sidebar, children }: DashboardPageLayoutProps) {
  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      {sidebar}
      <div className="flex min-w-0 flex-1 flex-col overflow-y-auto p-6">
        {children}
      </div>
    </div>
  )
}

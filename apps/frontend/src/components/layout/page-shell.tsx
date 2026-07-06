import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export type PageShellLayout = 'stack' | 'split' | 'fill'

export interface PageShellProps {
  children: ReactNode
  layout?: PageShellLayout
  className?: string
  actions?: ReactNode
  leading?: ReactNode
  description?: ReactNode
  stats?: ReactNode
  sidebar?: ReactNode
  contentClassName?: string
}

export function PageShell({
  children,
  layout = 'stack',
  className,
  actions,
  leading,
  description,
  stats,
  sidebar,
  contentClassName,
}: PageShellProps) {
  if (layout === 'split') {
    return (
      <div className={cn('flex min-h-0 flex-1 gap-4', className)}>
        {sidebar}
        <div className={cn('flex min-w-0 flex-1 flex-col gap-4', contentClassName)}>{children}</div>
      </div>
    )
  }

  if (layout === 'fill') {
    return <div className={cn('flex min-h-0 flex-1 flex-col', className)}>{children}</div>
  }

  const hasToolbar = leading || actions

  return (
    <div className={cn('space-y-6', className)}>
      {description}
      {hasToolbar && (
        <div className="flex items-center justify-between gap-4">
          {leading ? <div className="min-w-0 flex-1">{leading}</div> : <div />}
          {actions ? <div className="flex shrink-0 items-center gap-3">{actions}</div> : null}
        </div>
      )}
      {stats}
      {children}
    </div>
  )
}

import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export interface PageShellProps {
  children: ReactNode
  className?: string
  testId?: string
}

export function PageShell({ children, className, testId }: PageShellProps) {
  return (
    <div className={cn('space-y-6', className)} data-testid={testId}>
      {children}
    </div>
  )
}

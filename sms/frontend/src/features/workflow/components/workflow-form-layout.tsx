import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface WorkflowFormLayoutProps {
  className?: string
  children: ReactNode
}

export function WorkflowFormLayout({ className, children }: WorkflowFormLayoutProps) {
  return <div className={cn('max-w-md space-y-6', className)}>{children}</div>
}

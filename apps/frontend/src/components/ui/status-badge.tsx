import type { ReactNode } from 'react'
import { Badge } from '@/components/ui/badge'
import { STATUS_BADGE_STYLES, type StatusBadgeVariant } from '@/lib/labels'
import { cn } from '@/lib/utils'

interface StatusBadgeProps {
  variant: StatusBadgeVariant
  children: ReactNode
  className?: string
}

export function StatusBadge({ variant, children, className }: StatusBadgeProps) {
  return (
    <Badge variant="outline" className={cn('border-0', STATUS_BADGE_STYLES[variant], className)}>
      {children}
    </Badge>
  )
}

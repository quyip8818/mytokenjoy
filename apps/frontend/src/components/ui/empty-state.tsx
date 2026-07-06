import type { LucideIcon } from 'lucide-react'
import { Inbox } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  actionClassName?: string
  actionId?: string
  icon?: LucideIcon
  compact?: boolean
  className?: string
}

export function EmptyState({
  title,
  description,
  actionLabel,
  onAction,
  actionClassName,
  actionId,
  icon: Icon = Inbox,
  compact = false,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-lg border border-dashed border-border/60 bg-muted/30 px-4 text-center',
        compact ? 'py-8' : 'py-12',
        className,
      )}
    >
      {!compact && (
        <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
          <Icon className="h-6 w-6 text-primary" />
        </div>
      )}
      <p className="text-sm font-medium text-foreground">{title}</p>
      {description && <p className="text-sm text-muted-foreground mt-1 max-w-sm">{description}</p>}
      {actionLabel && onAction && (
        <Button
          id={actionId}
          size="sm"
          variant="brand"
          className={cn('mt-4', actionClassName)}
          onClick={onAction}
        >
          {actionLabel}
        </Button>
      )}
    </div>
  )
}

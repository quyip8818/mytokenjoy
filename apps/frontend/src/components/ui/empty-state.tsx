import type { LucideIcon } from 'lucide-react'
import { Inbox } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export interface EmptyStateProps {
  variant?: 'prominent' | 'inline' | 'minimal'
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  actionClassName?: string
  actionId?: string
  icon?: LucideIcon
  className?: string
}

export function EmptyState({
  variant = 'prominent',
  title,
  description,
  actionLabel,
  onAction,
  actionClassName,
  actionId,
  icon: Icon = Inbox,
  className,
}: EmptyStateProps) {
  if (variant === 'minimal') {
    return (
      <div className={cn('flex h-full flex-col items-center justify-center gap-2 p-8', className)}>
        <Icon className="size-8 text-muted-foreground/40" strokeWidth={1.5} />
        <p className="text-sm text-muted-foreground">{title}</p>
      </div>
    )
  }

  if (variant === 'inline') {
    return (
      <div className={cn('flex items-center gap-2 py-6 text-sm text-muted-foreground', className)}>
        <Icon className="size-4 text-muted-foreground/60" />
        <span>{title}</span>
        {description && <span className="text-muted-foreground/60">— {description}</span>}
      </div>
    )
  }

  // prominent (default)
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-lg border border-dashed border-border/60 bg-muted/30 px-4 py-12 text-center',
        className,
      )}
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
        <Icon className="h-6 w-6 text-primary" />
      </div>
      <p className="text-sm font-medium text-foreground">{title}</p>
      {description && <p className="mt-1 max-w-sm text-sm text-muted-foreground">{description}</p>}
      {actionLabel && onAction && (
        <Button
          id={actionId}
         
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

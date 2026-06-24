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
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-12 px-4 text-center rounded-lg border border-dashed border-border/60 bg-slate-50/30',
        className,
      )}
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-indigo-50 mb-4">
        <Icon className="h-6 w-6 text-indigo-500" />
      </div>
      <p className="text-sm font-medium text-foreground">{title}</p>
      {description && <p className="text-sm text-muted-foreground mt-1 max-w-sm">{description}</p>}
      {actionLabel && onAction && (
        <Button
          id={actionId}
          size="sm"
          className={cn(
            'mt-4 bg-gradient-to-r from-indigo-600 to-violet-600 text-white',
            actionClassName,
          )}
          onClick={onAction}
        >
          {actionLabel}
        </Button>
      )}
    </div>
  )
}

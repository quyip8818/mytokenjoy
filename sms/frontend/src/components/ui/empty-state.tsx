import { cn } from '@/lib/utils'

export interface EmptyStateProps {
  message?: string
  className?: string
}

export function EmptyState({ message = '暂无数据', className }: EmptyStateProps) {
  return (
    <div className={cn('flex items-center justify-center py-12 text-muted-foreground', className)}>
      <p>{message}</p>
    </div>
  )
}

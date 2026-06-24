import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PageLoadingProps {
  label?: string
  className?: string
}

export function PageLoading({ label = '加载中...', className }: PageLoadingProps) {
  return (
    <div
      className={cn(
        'flex items-center justify-center gap-2 py-16 text-sm text-muted-foreground',
        className,
      )}
    >
      <Loader2 className="h-4 w-4 animate-spin text-primary" />
      {label}
    </div>
  )
}

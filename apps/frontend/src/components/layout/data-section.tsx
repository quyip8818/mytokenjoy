import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { PageLoading } from '@/components/ui/page-loading'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { cn } from '@/lib/utils'

export interface DataSectionEmptyProps {
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  icon?: LucideIcon
}

export type DataSectionLoadingVariant = 'spinner' | 'skeleton'

export interface DataSectionProps {
  title?: string
  headerAction?: ReactNode
  loading?: boolean
  loadingVariant?: DataSectionLoadingVariant
  skeletonRows?: number
  skeletonColumns?: number
  empty?: DataSectionEmptyProps | null
  error?: Error | null
  onRetry?: () => void
  children: ReactNode
  className?: string
  contentClassName?: string
}

export function DataSection({
  title,
  headerAction,
  loading = false,
  loadingVariant = 'skeleton',
  skeletonRows = 5,
  skeletonColumns = 6,
  empty = null,
  error = null,
  onRetry,
  children,
  className,
  contentClassName,
}: DataSectionProps) {
  const hasHeader = title || headerAction

  return (
    <Card className={cn('border-transparent shadow-card', className)}>
      <CardContent className={cn('pt-5 pb-4', contentClassName)}>
        {hasHeader && (
          <div className="mb-4 flex items-center justify-between gap-3">
            {title ? (
              <h3 className="text-sm font-semibold text-foreground/80">{title}</h3>
            ) : (
              <div />
            )}
            {headerAction}
          </div>
        )}
        {loading ? (
          loadingVariant === 'spinner' ? (
            <PageLoading />
          ) : (
            <TableSkeleton rows={skeletonRows} columns={skeletonColumns} />
          )
        ) : error ? (
          <ErrorState message={error.message} onRetry={onRetry} />
        ) : empty ? (
          <EmptyState
            icon={empty.icon}
            title={empty.title}
            description={empty.description}
            actionLabel={empty.actionLabel}
            onAction={empty.onAction}
          />
        ) : (
          children
        )}
      </CardContent>
    </Card>
  )
}

import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { PageLoading } from '@/components/ui/page-loading'
import { TableSkeleton } from '@/components/ui/table-skeleton'

export interface DataSectionEmptyProps {
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  icon?: LucideIcon
  variant?: 'prominent' | 'inline' | 'minimal'
}

export type DataSectionLoadingVariant = 'spinner' | 'skeleton'

export interface DataSectionProps {
  loading?: boolean
  loadingVariant?: DataSectionLoadingVariant
  skeletonRows?: number
  skeletonColumns?: number
  error?: Error | null
  onRetry?: () => void
  empty?: DataSectionEmptyProps | null
  children: ReactNode
  className?: string
}

/**
 * Pure state-switching layer. Does NOT render any container decoration.
 * Wrap with Card externally if you need borders/shadow.
 */
export function DataSection({
  loading = false,
  loadingVariant = 'skeleton',
  skeletonRows = 5,
  skeletonColumns = 6,
  error = null,
  onRetry,
  empty = null,
  children,
  className,
}: DataSectionProps) {
  if (loading) {
    return loadingVariant === 'spinner' ? (
      <PageLoading className={className} />
    ) : (
      <div className={className}>
        <TableSkeleton rows={skeletonRows} columns={skeletonColumns} />
      </div>
    )
  }

  if (error) {
    return <ErrorState message={error.message} onRetry={onRetry} className={className} />
  }

  if (empty) {
    return (
      <EmptyState
        variant={empty.variant}
        icon={empty.icon}
        title={empty.title}
        description={empty.description}
        actionLabel={empty.actionLabel}
        onAction={empty.onAction}
        className={className}
      />
    )
  }

  return className ? <div className={className}>{children}</div> : <>{children}</>
}

import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'
import { DataSection, type DataSectionEmptyProps } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'

interface AuditFilteredPageProps<T> {
  title: string
  actions: ReactNode
  loading: boolean
  items: readonly T[]
  empty: DataSectionEmptyProps
  children: ReactNode
}

export function AuditFilteredPage<T>({
  title,
  actions,
  loading,
  items,
  empty,
  children,
}: AuditFilteredPageProps<T>) {
  return (
    <PageShell actions={actions}>
      <DataSection
        title={title}
        loading={loading}
        loadingVariant="spinner"
        empty={listEmpty(loading, items, empty)}
      >
        {children}
      </DataSection>
    </PageShell>
  )
}

export type AuditEmptyConfig = DataSectionEmptyProps & { icon: LucideIcon }

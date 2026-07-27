import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { useMyCallLogsPage } from '@/features/mydashboard'
import { CallLogsList } from './call-logs-list'

type MyCallLogsPageShellProps = ReturnType<typeof useMyCallLogsPage>

export function MyCallLogsPageShell({
  logs,
  total,
  page,
  totalPages,
  loading,
  error,
  refresh,
  setPage,
}: MyCallLogsPageShellProps) {
  return (
    <PageShell>
      <PageHeader title="我的用量" />

      <DataSection loading={loading} error={error} onRetry={() => void refresh()}>
        <CallLogsList
          logs={logs}
          total={total}
          page={page}
          totalPages={totalPages}
          onPageChange={setPage}
        />
      </DataSection>
    </PageShell>
  )
}

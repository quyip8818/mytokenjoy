import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useMemberCallLogsPage } from '@/features/member'
import { CallLogsList } from './call-logs-list'

type MemberCallLogsPageShellProps = ReturnType<typeof useMemberCallLogsPage>

export function MemberCallLogsPageShell({
  logs,
  total,
  page,
  totalPages,
  loading,
  error,
  refresh,
  setPage,
}: MemberCallLogsPageShellProps) {
  return (
    <PageShell
      description={<h1 className="text-sm font-semibold">使用记录</h1>}
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="p-0"
        className="overflow-hidden"
      >
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

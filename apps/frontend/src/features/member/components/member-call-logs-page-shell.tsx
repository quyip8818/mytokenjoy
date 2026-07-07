import { Link } from 'react-router'
import { ArrowLeft } from 'lucide-react'
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
      description={
        <div className="flex items-center gap-3">
          <Link to="/me" className="text-xs text-muted-foreground hover:text-foreground">
            <ArrowLeft className="size-4" />
          </Link>
          <h1 className="text-sm font-semibold">使用记录</h1>
        </div>
      }
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

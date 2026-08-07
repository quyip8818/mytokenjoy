import { ClipboardList } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { StatusBadge } from '@/components/ui/status-badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { listEmpty } from '@/lib/list-empty'
import type { useApprovalPage } from '../hooks/use-approval-page'
import { ApprovalTable } from './approval-table'

type ApprovalPageShellProps = ReturnType<typeof useApprovalPage>

export function ApprovalPageShell({
  approvals,
  loading,
  error,
  refresh,
  tab,
  setTab,
  pendingCount,
  handleApprove,
  handleReject,
  handleRetry,
}: ApprovalPageShellProps) {
  return (
    <PageShell>
      <PageHeader testId="page-approval" title="审批中心" />

      <Tabs value={tab} onValueChange={(value) => setTab(value as typeof tab)}>
        <TabsList>
          <TabsTrigger value="pending">
            待审批
            {tab === 'pending' && pendingCount > 0 && (
              <StatusBadge variant="info" className="ml-1.5">
                {pendingCount}
              </StatusBadge>
            )}
          </TabsTrigger>
          <TabsTrigger value="approved">已通过</TabsTrigger>
          <TabsTrigger value="rejected">已拒绝</TabsTrigger>
          <TabsTrigger value="all">全部</TabsTrigger>
        </TabsList>

        <TabsContent value={tab} className="mt-4">
          <Card className="border-border shadow-xs">
            <CardContent className="px-5 pt-5 pb-4">
              <DataSection
                loading={loading}
                error={error}
                onRetry={refresh}
                skeletonColumns={8}
                empty={listEmpty(loading, approvals, {
                  icon: ClipboardList,
                  title: '暂无审批',
                  description: '当前筛选条件下没有审批记录',
                  variant: 'inline',
                })}
              >
                <ApprovalTable
                  approvals={approvals}
                  onApprove={handleApprove}
                  onReject={handleReject}
                  onRetry={handleRetry}
                />
              </DataSection>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </PageShell>
  )
}

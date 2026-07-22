import { ClipboardList } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
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
  canResolve,
  pendingCount,
  handleApprove,
  handleReject,
  handleRetry,
}: ApprovalPageShellProps) {
  return (
    <PageShell>
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
            <CardContent className="pt-5 pb-4">
              <h3 className="mb-4 text-sm font-semibold text-foreground/80">审批列表</h3>
              <DataSection
                loading={loading}
                error={error}
                onRetry={refresh}
                skeletonColumns={8}
                className="border-0 shadow-none"
                contentClassName="p-0"
                empty={listEmpty(loading, approvals, {
                  icon: ClipboardList,
                  title: '暂无审批',
                  description: '当前筛选条件下没有审批记录',
                })}
              >
                <ApprovalTable
                  approvals={approvals}
                  canResolve={canResolve}
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

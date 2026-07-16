import { ClipboardList } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import type { useApprovalPage } from '@/features/keys'
import { ApprovalTable } from './approval-table'

type ApprovalPageShellProps = ReturnType<typeof useApprovalPage>

export function ApprovalPageShell({
  approvals,
  loading,
  error,
  refresh,
  tab,
  setTab,
  canApprove,
  pendingCount,
  rowClass,
  handleApprove,
  handleReject,
  openSubmit,
}: ApprovalPageShellProps) {
  return (
    <PageShell
      actions={
        <PermissionGate permission={PERMISSION.SELF_APPROVAL}>
          <Button variant="brand" onClick={openSubmit}>
            发起申请
          </Button>
        </PermissionGate>
      }
    >
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
              <h3 className="mb-4 text-sm font-semibold text-foreground/80">申请列表</h3>
              <DataSection
                loading={loading}
                error={error}
                onRetry={refresh}
                skeletonColumns={7}
                className="border-0 shadow-none"
                contentClassName="p-0"
                empty={listEmpty(loading, approvals, {
                  icon: ClipboardList,
                  title: '暂无申请',
                  description: '当前筛选条件下没有审批记录',
                })}
              >
                <ApprovalTable
                  approvals={approvals}
                  canApprove={canApprove}
                  rowClass={rowClass}
                  onApprove={handleApprove}
                  onReject={handleReject}
                />
              </DataSection>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </PageShell>
  )
}

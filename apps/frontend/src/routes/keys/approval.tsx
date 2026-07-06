import { ClipboardList } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { ApprovalTable } from '@/routes/keys/components/approval-table'
import { useApprovalPage } from '@/routes/keys/hooks/use-approval-page'

export default function ApprovalPage() {
  const {
    approvals,
    loading,
    error,
    refresh,
    tab,
    setTab,
    canSubmit,
    pendingCount,
    hasKeyType,
    hasQuotaType,
    rowClass,
    handleRowClick,
    openSubmit,
  } = useApprovalPage()

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
      <Tabs value={tab} onValueChange={(v) => setTab(v as typeof tab)}>
        <TabsList>
          <TabsTrigger value="pending">
            待我审批
            {tab === 'pending' && pendingCount > 0 && (
              <StatusBadge variant="info" className="ml-1.5">
                {pendingCount}
              </StatusBadge>
            )}
          </TabsTrigger>
          <TabsTrigger value="mine">我的申请</TabsTrigger>
          <TabsTrigger value="all">全部</TabsTrigger>
        </TabsList>

        <TabsContent value={tab} className="mt-4">
          <DataSection
            loading={loading}
            error={error}
            onRetry={refresh}
            skeletonColumns={7}
            empty={listEmpty(loading, approvals, {
              icon: ClipboardList,
              title: '暂无审批记录',
              description:
                tab === 'pending' ? '当前没有待处理的审批申请' : '发起申请后记录将显示在这里',
              actionLabel: canSubmit ? '发起申请' : undefined,
              onAction: canSubmit ? openSubmit : undefined,
            })}
          >
            <ApprovalTable
              approvals={approvals}
              hasKeyType={hasKeyType}
              hasQuotaType={hasQuotaType}
              rowClass={rowClass}
              onRowClick={handleRowClick}
            />
          </DataSection>
        </TabsContent>
      </Tabs>
    </PageShell>
  )
}

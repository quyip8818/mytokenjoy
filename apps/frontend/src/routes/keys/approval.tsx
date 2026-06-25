import { ClipboardList } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { ApprovalStatusBadge } from '@/lib/label-badges'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
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
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead>类型</TableHead>
                  <TableHead>申请人</TableHead>
                  <TableHead>部门</TableHead>
                  <TableHead>申请理由</TableHead>
                  {hasKeyType && <TableHead>申请模型</TableHead>}
                  {hasQuotaType && <TableHead className="text-right">额度 (¥)</TableHead>}
                  <TableHead>状态</TableHead>
                  <TableHead>申请时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {approvals.map((a) => (
                  <TableRow
                    key={a.id}
                    className={`cursor-pointer ${rowClass(a.id)}`}
                    onClick={() => handleRowClick(a)}
                  >
                    <TableCell>
                      <StatusBadge variant="neutral">
                        {a.type === 'key' ? 'Key 申请' : '额度追加'}
                      </StatusBadge>
                    </TableCell>
                    <TableCell className="font-medium">{a.applicant}</TableCell>
                    <TableCell className="text-muted-foreground">{a.department}</TableCell>
                    <TableCell className="max-w-48 truncate text-sm">{a.reason}</TableCell>
                    {hasKeyType && (
                      <TableCell className="text-sm text-muted-foreground">
                        {a.type === 'key' ? a.requestedModels.join(', ') || '—' : '—'}
                      </TableCell>
                    )}
                    {hasQuotaType && (
                      <TableCell className="text-right">
                        {a.type === 'quota' ? a.requestedQuota.toLocaleString() : '—'}
                      </TableCell>
                    )}
                    <TableCell>
                      <ApprovalStatusBadge status={a.status} />
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{a.createdAt}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </DataSection>
        </TabsContent>
      </Tabs>
    </PageShell>
  )
}

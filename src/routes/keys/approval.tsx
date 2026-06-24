import { useCallback, useEffect, useMemo, useState } from 'react'
import { ClipboardList } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ApprovalStatusBadge } from '@/lib/label-badges'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { approvalApi } from '@/api/keys'
import type { KeyApproval } from '@/api/types'
import { useDemoRole } from '@/features/demo'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { EmptyState } from '@/components/ui/empty-state'
import { useRowHighlight } from '@/lib/use-row-highlight'

export default function ApprovalPage() {
  const { memberId } = useDemoRole()
  const { open } = useWorkflow()
  const { flashRow, rowClass } = useRowHighlight()
  const [approvals, setApprovals] = useState<KeyApproval[]>([])
  const [tab, setTab] = useState<'pending' | 'mine' | 'all'>('pending')

  const load = useCallback(async () => {
    const data = await approvalApi.list({
      tab: tab === 'all' ? undefined : tab,
      memberId: tab === 'mine' ? memberId : undefined,
    })
    setApprovals(data)
  }, [tab, memberId])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      const data = await approvalApi.list({
        tab: tab === 'all' ? undefined : tab,
        memberId: tab === 'mine' ? memberId : undefined,
      })
      if (!cancelled) setApprovals(data)
    })()
    return () => {
      cancelled = true
    }
  }, [tab, memberId])

  const pendingCount = approvals.filter((a) => a.status === 'pending').length

  const hasKeyType = useMemo(() => approvals.some((a) => a.type === 'key'), [approvals])
  const hasQuotaType = useMemo(() => approvals.some((a) => a.type === 'quota'), [approvals])

  const getTypeBadge = (type: string) => (
    <Badge variant="outline" className="text-xs">
      {type === 'key' ? 'Key 申请' : '额度追加'}
    </Badge>
  )

  const handleRowClick = (approval: KeyApproval) => {
    open('approval-review', {
      approval,
      onSuccess: () => {
        void load()
        flashRow(approval.id)
      },
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end">
        <Button
          className="bg-gradient-to-r from-indigo-600 to-violet-600 text-white shadow-button hover:from-indigo-500 hover:to-violet-500"
          onClick={() => open('approval-submit', { onSuccess: load })}
        >
          发起申请
        </Button>
      </div>

      <Tabs value={tab} onValueChange={(v) => setTab(v as typeof tab)}>
        <TabsList>
          <TabsTrigger value="pending">
            待我审批
            {tab === 'pending' && pendingCount > 0 && (
              <span className="ml-1.5 inline-flex items-center justify-center rounded-full bg-indigo-100 px-1.5 py-0.5 text-xs font-medium text-indigo-700">
                {pendingCount}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="mine">我的申请</TabsTrigger>
          <TabsTrigger value="all">全部</TabsTrigger>
        </TabsList>

        <TabsContent value={tab} className="mt-4">
          <Card className="border-border/50 shadow-card">
            <CardContent className="pt-5 pb-4">
              {approvals.length === 0 ? (
                <EmptyState
                  icon={ClipboardList}
                  title="暂无审批记录"
                  description={
                    tab === 'pending' ? '当前没有待处理的审批申请' : '发起申请后记录将显示在这里'
                  }
                  actionLabel="发起申请"
                  onAction={() => open('approval-submit', { onSuccess: load })}
                />
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow className="border-border/50 hover:bg-transparent">
                      <TableHead className="text-xs font-semibold text-muted-foreground">
                        类型
                      </TableHead>
                      <TableHead className="text-xs font-semibold text-muted-foreground">
                        申请人
                      </TableHead>
                      <TableHead className="text-xs font-semibold text-muted-foreground">
                        部门
                      </TableHead>
                      <TableHead className="text-xs font-semibold text-muted-foreground">
                        申请理由
                      </TableHead>
                      {hasKeyType && (
                        <TableHead className="text-xs font-semibold text-muted-foreground">
                          申请模型
                        </TableHead>
                      )}
                      {hasQuotaType && (
                        <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                          额度 (¥)
                        </TableHead>
                      )}
                      <TableHead className="text-xs font-semibold text-muted-foreground">
                        状态
                      </TableHead>
                      <TableHead className="text-xs font-semibold text-muted-foreground">
                        申请时间
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {approvals.map((a) => (
                      <TableRow
                        key={a.id}
                        className={`cursor-pointer ${rowClass(a.id)}`}
                        onClick={() => handleRowClick(a)}
                      >
                        <TableCell>{getTypeBadge(a.type)}</TableCell>
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
                        <TableCell className="text-sm text-muted-foreground">
                          {a.createdAt}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

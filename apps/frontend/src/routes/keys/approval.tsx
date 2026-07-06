import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { approvalApi } from '@/api/keys'
import type { KeyApproval } from '@/api/types'

export default function ApprovalPage() {
  const [approvals, setApprovals] = useState<KeyApproval[]>([])
  const [tab, setTab] = useState('pending')

  useEffect(() => {
    approvalApi.list().then(setApprovals)
  }, [])

  const filtered = approvals.filter(a => tab === 'all' || a.status === tab)
  const pendingCount = approvals.filter(a => a.status === 'pending').length

  const handleApprove = async (id: string) => {
    await approvalApi.approve(id)
    setApprovals(approvals.map(a => a.id === id ? { ...a, status: 'approved', approver: '当前用户', resolvedAt: new Date().toISOString() } : a))
  }

  const handleReject = async (id: string) => {
    await approvalApi.reject(id)
    setApprovals(approvals.map(a => a.id === id ? { ...a, status: 'rejected', approver: '当前用户', resolvedAt: new Date().toISOString() } : a))
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-amber-50 text-amber-700">待审批</span>
      case 'approved':
        return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700">已通过</span>
      case 'rejected':
        return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-700">已拒绝</span>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  return (
    <div className="space-y-8">
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="pending">
            待审批
            {pendingCount > 0 && (
              <span className="ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">
                {pendingCount}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="approved">已通过</TabsTrigger>
          <TabsTrigger value="rejected">已拒绝</TabsTrigger>
          <TabsTrigger value="all">全部</TabsTrigger>
        </TabsList>

        <TabsContent value={tab} className="mt-4">
          <Card className="shadow-xs border-border">
            <CardContent className="pt-5 pb-4">
              <h3 className="text-sm font-semibold text-foreground/80 mb-4">申请列表</h3>
              <Table>
                <TableHeader>
                  <TableRow className="border-border/50 hover:bg-transparent">
                    <TableHead className="text-xs font-semibold text-muted-foreground">申请人</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground">申请理由</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground text-right">额度 (¥)</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground">模型</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground">状态</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground">申请时间</TableHead>
                    <TableHead className="text-xs font-semibold text-muted-foreground">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.length === 0 ? (
                    <TableRow><TableCell colSpan={8} className="text-center text-muted-foreground py-8">暂无数据</TableCell></TableRow>
                  ) : filtered.map((a) => (
                    <TableRow key={a.id} className="border-border-subtle hover:bg-muted/50">
                      <TableCell className="font-medium">{a.applicant}</TableCell>
                      <TableCell className="text-muted-foreground">{a.department}</TableCell>
                      <TableCell className="max-w-48 truncate text-sm">{a.reason}</TableCell>
                      <TableCell className="text-right">{a.requestedQuota.toLocaleString()}</TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {a.requestedModels.map(m => <Badge key={m} variant="outline" className="text-xs bg-primary/10 text-primary border-primary/20">{m}</Badge>)}
                        </div>
                      </TableCell>
                      <TableCell>{getStatusBadge(a.status)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{a.createdAt}</TableCell>
                      <TableCell>
                        {a.status === 'pending' && (
                          <div className="flex gap-1">
                            <Button
                              size="sm"
                              onClick={() => handleApprove(a.id)}
                              className="bg-primary text-primary-foreground hover:bg-primary/90"
                            >
                              通过
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              className="text-red-600 border-red-200 hover:bg-red-50"
                              onClick={() => handleReject(a.id)}
                            >
                              拒绝
                            </Button>
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

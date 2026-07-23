import type { ApprovalRequest } from '@/api/types'
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
import { formatDisplayCurrency } from '@/lib/quota-display'

const TYPE_LABELS: Record<string, string> = {
  key: 'Key 申请',
  member_budget: '额度追加',
  project_budget: '项目预算',
  project_member_budget: '项目成员额度',
}

const STATUS_STYLES: Record<string, string> = {
  pending: 'bg-amber-50 text-amber-700',
  approved: 'bg-emerald-50 text-emerald-700',
  rejected: 'bg-red-50 text-red-700',
  cancelled: 'bg-gray-50 text-gray-600',
  failed: 'bg-orange-50 text-orange-700',
}

const STATUS_LABELS: Record<string, string> = {
  pending: '待审批',
  approved: '已通过',
  rejected: '已拒绝',
  cancelled: '已撤回',
  failed: '执行失败',
}

function getDisplayAmount(approval: ApprovalRequest): number | null {
  const meta = approval.metadata
  if ('requestedBudget' in meta && typeof meta.requestedBudget === 'number') {
    return meta.requestedBudget
  }
  if ('amount' in meta && typeof meta.amount === 'number') {
    return meta.amount
  }
  return null
}

function getReason(approval: ApprovalRequest): string {
  const meta = approval.metadata
  return typeof meta.reason === 'string' ? meta.reason : ''
}

function getDepartmentName(approval: ApprovalRequest): string {
  const meta = approval.metadata
  return typeof meta.departmentName === 'string' ? meta.departmentName : ''
}

interface ApprovalTableProps {
  approvals: ApprovalRequest[]
  onApprove: (id: string) => void
  onReject: (id: string, reason: string) => void
  onRetry?: (id: string) => void
}

export function ApprovalTable({
  approvals,
  onApprove,
  onReject,
  onRetry,
}: ApprovalTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            类型
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            申请人
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            部门
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            申请理由
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            金额
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            状态
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            时间
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            操作
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {approvals.map((approval) => (
          <TableRow key={approval.id} className="even:bg-muted/40">
            <TableCell>
              <Badge variant="outline" className="text-xs">
                {TYPE_LABELS[approval.type] ?? approval.type}
              </Badge>
            </TableCell>
            <TableCell className="font-medium">{approval.applicantName}</TableCell>
            <TableCell className="text-muted-foreground">{getDepartmentName(approval)}</TableCell>
            <TableCell className="max-w-48 truncate text-sm">{getReason(approval)}</TableCell>
            <TableCell className="text-right tabular-nums">
              {getDisplayAmount(approval) != null
                ? formatDisplayCurrency(getDisplayAmount(approval)!)
                : '—'}
            </TableCell>
            <TableCell>
              <Badge variant="outline" className={STATUS_STYLES[approval.status] ?? ''}>
                {STATUS_LABELS[approval.status] ?? approval.status}
              </Badge>
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">{approval.createdAt}</TableCell>
            <TableCell>
              <ApprovalActions
                approval={approval}
                onApprove={onApprove}
                onReject={onReject}
                onRetry={onRetry}
              />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function ApprovalActions({
  approval,
  onApprove,
  onReject,
  onRetry,
}: {
  approval: ApprovalRequest
  onApprove: (id: string) => void
  onReject: (id: string, reason: string) => void
  onRetry?: (id: string) => void
}) {
  if (approval.status === 'pending' && approval.canResolve) {
    return (
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          className="h-8 text-emerald-700 hover:text-emerald-800"
          onClick={() => onApprove(approval.id)}
        >
          通过
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 text-red-600 hover:text-red-700"
          onClick={() => onReject(approval.id, '已拒绝')}
        >
          拒绝
        </Button>
      </div>
    )
  }
  if (approval.status === 'failed' && approval.canResolve && onRetry) {
    return (
      <Button
        variant="ghost"
        size="sm"
        className="h-8 text-orange-700 hover:text-orange-800"
        onClick={() => onRetry(approval.id)}
      >
        重试
      </Button>
    )
  }
  return <span className="text-sm text-muted-foreground">—</span>
}

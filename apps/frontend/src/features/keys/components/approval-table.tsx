import type { KeyApproval } from '@/api/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { ApprovalStatusBadge } from './status-badges'
import { currencySymbol, formatDisplayCurrency } from '@/lib/points'
import { useBillingExchange } from '@/features/session'

interface ApprovalTableProps {
  approvals: KeyApproval[]
  canApprove: boolean
  rowClass: (id: string) => string | undefined
  onApprove: (id: string) => void
  onReject: (id: string) => void
}

export function ApprovalTable({
  approvals,
  canApprove,
  rowClass,
  onApprove,
  onReject,
}: ApprovalTableProps) {
  const { billingCurrency } = useBillingExchange()
  const currencyLabel = currencySymbol(billingCurrency)

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
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
            申请额度 ({currencyLabel})
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            状态
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            申请时间
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            操作
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {approvals.map((approval) => (
          <TableRow key={approval.id} className={`even:bg-muted/40 ${rowClass(approval.id)}`}>
            <TableCell className="font-medium">{approval.applicant}</TableCell>
            <TableCell className="text-muted-foreground">{approval.department}</TableCell>
            <TableCell className="max-w-48 truncate text-sm">{approval.reason}</TableCell>
            <TableCell className="text-right tabular-nums">
              {formatDisplayCurrency(approval.requestedBudget)}
            </TableCell>
            <TableCell>
              <ApprovalStatusBadge status={approval.status} />
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">{approval.createdAt}</TableCell>
            <TableCell>
              {canApprove && approval.status === 'pending' ? (
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
                    onClick={() => onReject(approval.id)}
                  >
                    拒绝
                  </Button>
                </div>
              ) : (
                <span className="text-sm text-muted-foreground">—</span>
              )}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

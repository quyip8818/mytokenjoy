import type { KeyApproval } from '@/api/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/ui/status-badge'
import { ApprovalStatusBadge } from '@/lib/label-badges'

interface ApprovalTableProps {
  approvals: KeyApproval[]
  hasKeyType: boolean
  hasQuotaType: boolean
  rowClass: (id: string) => string | undefined
  onRowClick: (approval: KeyApproval) => void
}

export function ApprovalTable({
  approvals,
  hasKeyType,
  hasQuotaType,
  rowClass,
  onRowClick,
}: ApprovalTableProps) {
  return (
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
            onClick={() => onRowClick(a)}
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
  )
}

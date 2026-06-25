import { Link } from 'react-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { ROUTES } from '@/config/routes'
import type { Department } from '@/api/types'

interface StructureSummaryCardProps {
  selectedDept?: Department
  total: number
  inactiveCount: number
  approvalPendingCount: number
  onBatchInvite: () => void
}

export function StructureSummaryCard({
  selectedDept,
  total,
  inactiveCount,
  approvalPendingCount,
  onBatchInvite,
}: StructureSummaryCardProps) {
  return (
    <Card size="sm" className="border-border/50 shadow-card">
      <CardContent className="flex items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-6">
          <span className="text-base font-semibold">{selectedDept?.name ?? '全部成员'}</span>
          <span className="text-sm text-muted-foreground">
            总人数: <span className="font-medium text-foreground">{total}</span>
          </span>
          {inactiveCount > 0 && (
            <StatusBadge variant="warning">未激活 {inactiveCount} 人</StatusBadge>
          )}
          {approvalPendingCount > 0 && (
            <PermissionGate permission={PERMISSION.BUDGET_APPROVE}>
              <Link to={ROUTES.keysApproval} className="text-sm text-blue-600 hover:text-blue-500">
                待审批申请: {approvalPendingCount} → 去审批
              </Link>
            </PermissionGate>
          )}
        </div>
        <PermissionGate write>
          {inactiveCount > 0 && (
            <Button size="sm" variant="outline" onClick={onBatchInvite}>
              批量发送邀请
            </Button>
          )}
        </PermissionGate>
      </CardContent>
    </Card>
  )
}

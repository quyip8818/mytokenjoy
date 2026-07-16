import { useState } from 'react'
import { toast } from 'sonner'
import type { KeyApproval } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '@/features/workflow'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { Badge } from '@/components/ui/badge'
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'
import { useModelLabels } from '@/features/models'
import { formatDisplayCurrency } from '@/lib/points'

export function ApprovalReviewWorkflow({
  entry,
  onPush,
  onClose,
}: WorkflowComponentProps<'approval-review'>) {
  const apis = useInjectedApis()
  const { labelFor } = useModelLabels(apis)
  const { closeAll } = useWorkflow()
  const approval = entry.payload.approval as KeyApproval
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [submitting, setSubmitting] = useState(false)

  const typeLabel = approval.type === 'key' ? 'Key 申请' : '额度追加'

  const handleApprove = async () => {
    if (approval.type === 'budget') {
      const check = await apis.approvalApi.checkBudget(approval.id)
      if (!check.sufficient) {
        onPush('budget-check', {
          reservedPool: check.reservedPool,
          requested: check.requested,
        })
        return
      }
    }
    setSubmitting(true)
    try {
      await apis.approvalApi.approve(approval.id)
      toast.success('已通过')
      onSuccess?.()
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '审批失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleReject = () => {
    onPush('reject-reason', {
      approvalId: approval.id,
      onSuccess,
    })
  }

  return (
    <WorkflowPanelChrome
      title="审批处理"
      onClose={onClose}
      contextBar={`申请人：${approval.applicant} · ${approval.department}`}
      footer={
        approval.status === 'pending' ? (
          <WorkflowPanelFooter
            onCancel={onClose}
            cancelLabel="关闭"
            destructiveLabel="拒绝"
            onDestructive={handleReject}
            primaryLabel={submitting ? '处理中...' : '通过'}
            onPrimary={handleApprove}
            primaryDisabled={submitting}
          />
        ) : (
          <WorkflowPanelFooter
            onCancel={onClose}
            cancelLabel="关闭"
            primaryLabel="关闭"
            onPrimary={onClose}
          />
        )
      }
    >
      <div className="grid grid-cols-5 gap-8">
        <div className="col-span-3 space-y-5">
          <div className="flex items-center gap-2">
            <Badge variant="outline">{typeLabel}</Badge>
            <Badge
              variant="outline"
              className={
                approval.status === 'pending'
                  ? 'bg-amber-50 text-amber-700'
                  : approval.status === 'approved'
                    ? 'bg-emerald-50 text-emerald-700'
                    : 'bg-red-50 text-red-700'
              }
            >
              {approval.status === 'pending'
                ? '待审批'
                : approval.status === 'approved'
                  ? '已通过'
                  : '已拒绝'}
            </Badge>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">申请理由</h4>
            <p className="text-sm">{approval.reason}</p>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-1">申请额度</h4>
            <p className="text-lg font-semibold">
              {formatDisplayCurrency(approval.requestedBudget)}
            </p>
          </div>
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-2">申请模型</h4>
            <div className="flex flex-wrap gap-1">
              {approval.requestedModels.map((modelId) => (
                <Badge key={modelId} variant="outline" className="text-xs">
                  {labelFor(modelId)}
                </Badge>
              ))}
            </div>
          </div>
          {approval.rejectReason && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-1">拒绝理由</h4>
              <p className="text-sm text-red-600">{approval.rejectReason}</p>
            </div>
          )}
        </div>
        <WorkflowInfoBox fullWidth className="space-y-3">
          <h4 className="font-semibold">申请信息</h4>
          <p className="text-muted-foreground">申请人：{approval.applicant}</p>
          <p className="text-muted-foreground">部门：{approval.department}</p>
          <p className="text-muted-foreground">申请时间：{approval.createdAt}</p>
          {approval.approver && (
            <p className="text-muted-foreground">审批人：{approval.approver}</p>
          )}
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}

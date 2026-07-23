import { useState } from 'react'
import { toast } from 'sonner'
import type { ApprovalRequest } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '@/features/workflow'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { Badge } from '@/components/ui/badge'
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'
import { useModelLabels } from '@/features/models'
import { formatDisplayCurrency } from '@/lib/quota-display'

const TYPE_LABELS: Record<string, string> = {
  key: 'Key 申请',
  member_budget: '额度追加',
  project_budget: '项目预算',
  project_member_budget: '项目成员额度',
}

const STATUS_LABELS: Record<string, string> = {
  pending: '待审批',
  approved: '已通过',
  rejected: '已拒绝',
  cancelled: '已撤回',
  failed: '执行失败',
}

const STATUS_STYLES: Record<string, string> = {
  pending: 'bg-amber-50 text-amber-700',
  approved: 'bg-emerald-50 text-emerald-700',
  rejected: 'bg-red-50 text-red-700',
  cancelled: 'bg-gray-50 text-gray-600',
  failed: 'bg-orange-50 text-orange-700',
}

function getMetaDepartmentName(approval: ApprovalRequest): string {
  const v = approval.metadata.departmentName
  return typeof v === 'string' ? v : ''
}

function getMetaProjectName(approval: ApprovalRequest): string {
  const v = approval.metadata.projectName
  return typeof v === 'string' ? v : ''
}

export function ApprovalReviewWorkflow({
  entry,
  onPush,
  onClose,
}: WorkflowComponentProps<'approval-review'>) {
  const apis = useInjectedApis()
  const { labelFor } = useModelLabels(apis)
  const { closeAll } = useWorkflow()
  const approval = entry.payload.approval as ApprovalRequest
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [submitting, setSubmitting] = useState(false)

  const meta = approval.metadata as Record<string, unknown>
  const reason = (meta.reason as string) ?? ''
  const requestedBudget = (meta.requestedBudget as number) ?? (meta.amount as number) ?? 0
  const requestedModels = (meta.requestedModels as string[]) ?? []

  const handleApprove = async () => {
    // Pre-check budget sufficiency before approving
    const check = await apis.approvalApi.preCheck(approval.id)
    if (!check.sufficient) {
      onPush('budget-check', {
        reservedPool: check.reservedPool,
        requested: check.requested,
      })
      return
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
      contextBar={`申请人：${approval.applicantName} · ${getMetaDepartmentName(approval)}`}
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
            <Badge variant="outline">{TYPE_LABELS[approval.type] ?? approval.type}</Badge>
            <Badge variant="outline" className={STATUS_STYLES[approval.status] ?? ''}>
              {STATUS_LABELS[approval.status] ?? approval.status}
            </Badge>
          </div>
          {reason && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-1">申请理由</h4>
              <p className="text-sm">{reason}</p>
            </div>
          )}
          {requestedBudget > 0 && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-1">申请额度</h4>
              <p className="text-lg font-semibold">{formatDisplayCurrency(requestedBudget)}</p>
            </div>
          )}
          {requestedModels.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-2">申请模型</h4>
              <div className="flex flex-wrap gap-1">
                {requestedModels.map((modelId) => (
                  <Badge key={modelId} variant="outline" className="text-xs">
                    {labelFor(modelId)}
                  </Badge>
                ))}
              </div>
            </div>
          )}
          {approval.rejectReason && (
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-1">拒绝理由</h4>
              <p className="text-sm text-red-600">{approval.rejectReason}</p>
            </div>
          )}
        </div>
        <WorkflowInfoBox fullWidth className="space-y-3">
          <h4 className="font-semibold">申请信息</h4>
          <p className="text-muted-foreground">申请人：{approval.applicantName}</p>
          {getMetaDepartmentName(approval) && (
            <p className="text-muted-foreground">部门：{getMetaDepartmentName(approval)}</p>
          )}
          {getMetaProjectName(approval) && (
            <p className="text-muted-foreground">项目：{getMetaProjectName(approval)}</p>
          )}
          <p className="text-muted-foreground">申请时间：{approval.createdAt}</p>
          {approval.approverName && (
            <p className="text-muted-foreground">审批人：{approval.approverName}</p>
          )}
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}

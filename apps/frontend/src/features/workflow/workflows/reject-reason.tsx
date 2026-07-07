import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { useWorkflow } from '../use-workflow'

export function RejectReasonWorkflow({
  entry,
  onPop,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'reject-reason'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const approvalId = entry.payload.approvalId as string
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleReject = async () => {
    if (!reason.trim()) {
      toast.error('请填写拒绝理由')
      return
    }
    setSubmitting(true)
    try {
      await apis.approvalApi.reject(approvalId, reason)
      toast.success('已拒绝')
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('操作失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="拒绝理由"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel={submitting ? '提交中...' : '确认拒绝'}
          onPrimary={handleReject}
          primaryDisabled={!reason.trim() || submitting}
        />
      }
    >
      <WorkflowFormLayout className="space-y-3">
        <Label>请填写拒绝理由（必填）</Label>
        <Textarea
          value={reason}
          onChange={(e) => {
            setReason(e.target.value)
            onSetDirty(true)
          }}
          rows={4}
          placeholder="请说明拒绝原因..."
        />
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

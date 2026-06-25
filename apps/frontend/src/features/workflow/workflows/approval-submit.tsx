import { useState } from 'react'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import { useApis } from '@/api/use-apis'
import { useDemoRole } from '@/features/demo'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '../use-workflow'
import { pushModelPicker, useMemberWhitelist } from '../use-member-whitelist'
import type { ApprovalType } from '@/api/types'
import { MODEL_NOT_IN_DEPT_MESSAGE } from '@/lib/dashboard-constants'

export function ApprovalSubmitWorkflow({
  entry,
  onClose,
  onPush,
  onSetDirty,
}: WorkflowComponentProps<'approval-submit'>) {
  const apis = useApis()
  const { closeAll } = useWorkflow()
  const { memberId } = useDemoRole()
  const { resolveWhitelist } = useMemberWhitelist()
  const defaultType = (entry.payload.defaultType as ApprovalType) ?? 'quota'
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [type, setType] = useState<ApprovalType>(defaultType)
  const [reason, setReason] = useState('')
  const [quota, setQuota] = useState('3000')
  const [models, setModels] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)

  const handlePickModels = () => {
    void pushModelPicker(onPush, resolveWhitelist, {
      selectedModels: models,
      onConfirm: setModels,
      onSetDirty,
    })
  }

  const validateModels = async (): Promise<boolean> => {
    if (type !== 'key' || models.length === 0) return true
    const allowed = await resolveWhitelist()
    if (!allowed?.length) return true
    const invalid = models.filter((m) => !allowed.includes(m))
    if (invalid.length > 0) {
      toast.error(MODEL_NOT_IN_DEPT_MESSAGE)
      return false
    }
    return true
  }

  const handleSubmit = async () => {
    if (!(await validateModels())) return
    setSubmitting(true)
    try {
      await apis.approvalApi.create({
        type,
        reason,
        requestedQuota: Number(quota),
        requestedModels: models,
        memberId,
      })
      toast.success('申请已提交')
      onSuccess?.()
      closeAll()
    } catch (err) {
      const message = err instanceof ApiError ? err.message : '提交失败'
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="发起申请"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '提交中...' : '提交申请'}
          onPrimary={handleSubmit}
          primaryDisabled={!reason.trim() || submitting || (type === 'key' && models.length === 0)}
        />
      }
    >
      <WorkflowFormLayout variant="wide">
        <div className="space-y-1.5">
          <Label>申请类型</Label>
          <Select
            value={type}
            onValueChange={(v) => {
              setType(v as ApprovalType)
              onSetDirty(true)
            }}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="key">Key 申请</SelectItem>
              <SelectItem value="quota">额度追加</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1.5">
          <Label>申请理由</Label>
          <Textarea
            value={reason}
            onChange={(e) => {
              setReason(e.target.value)
              onSetDirty(true)
            }}
            rows={3}
            placeholder="请说明申请原因..."
          />
        </div>
        <div className="space-y-1.5">
          <Label>申请额度 (¥)</Label>
          <Input
            type="number"
            value={quota}
            onChange={(e) => {
              setQuota(e.target.value)
              onSetDirty(true)
            }}
          />
        </div>
        {type === 'key' && (
          <div className="space-y-3">
            <Label>申请模型</Label>
            <Button variant="outline" onClick={handlePickModels}>
              选择模型 {models.length > 0 && `(${models.length})`}
            </Button>
            {models.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {models.map((m) => (
                  <Badge key={m} variant="outline" className="text-xs">
                    {m}
                  </Badge>
                ))}
              </div>
            )}
          </div>
        )}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

import { useState } from 'react'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import { useInjectedApis } from '@/api/use-apis'
import { useSession } from '@/features/session'
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
import { MODEL_NOT_IN_DEPT_MESSAGE } from '@/features/dashboard'
import { useModelLabels } from '@/features/models/hooks/use-model-labels'

export function ApprovalSubmitWorkflow({
  entry,
  onClose,
  onPush,
  onSetDirty,
}: WorkflowComponentProps<'approval-submit'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const { memberId } = useSession()
  const { resolveAllowedModelIds } = useMemberWhitelist()
  const { labelFor } = useModelLabels(apis)
  const defaultType = (entry.payload.defaultType as ApprovalType) ?? 'budget'
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [type, setType] = useState<ApprovalType>(defaultType)
  const [reason, setReason] = useState('')
  const [requestedBudget, setRequestedBudget] = useState('3000')
  const [models, setModels] = useState<number[]>([])
  const [submitting, setSubmitting] = useState(false)

  const handlePickModels = () => {
    void pushModelPicker(onPush, resolveAllowedModelIds, {
      selectedModelIds: models,
      onConfirm: setModels,
      onSetDirty,
    })
  }

  const validateModels = async (): Promise<boolean> => {
    if (type !== 'key' || models.length === 0) return true
    const allowed = await resolveAllowedModelIds()
    if (!allowed?.length) {
      toast.error('无法加载可用模型白名单')
      return false
    }
    const allowedSet = new Set(allowed)
    const invalid = models.filter((id) => !allowedSet.has(id))
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
        requestedBudget: Number(requestedBudget),
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
              <SelectItem value="budget">额度追加</SelectItem>
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
            value={requestedBudget}
            onChange={(e) => {
              setRequestedBudget(e.target.value)
              onSetDirty(true)
            }}
          />
        </div>
        {type === 'key' && (
          <div className="space-y-3">
            <Label>申请模型</Label>
            <Button variant="outline" onClick={handlePickModels}>
              选择模型 ({models.length})
            </Button>
            <div className="flex flex-wrap gap-1">
              {models.map((modelId) => (
                <Badge key={modelId} variant="outline" className="text-xs">
                  {labelFor(modelId)}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

import { useState } from 'react'
import { toast } from 'sonner'
import { approvalApi } from '@/api/keys'
import { useDemoRole } from '@/features/demo'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
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

export function ApprovalSubmitWorkflow({
  entry,
  onClose,
  onPush,
  onSetDirty,
}: WorkflowComponentProps) {
  const { closeAll } = useWorkflow()
  const { memberId } = useDemoRole()
  const { resolveWhitelist } = useMemberWhitelist()
  const defaultType = (entry.payload.defaultType as ApprovalType) ?? 'quota'
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [type, setType] = useState<ApprovalType>(defaultType)
  const [reason, setReason] = useState('')
  const [quota, setQuota] = useState('3000')
  const [models, setModels] = useState<string[]>(['gpt-4o'])
  const [submitting, setSubmitting] = useState(false)

  const handlePickModels = () => {
    void pushModelPicker(onPush, resolveWhitelist, {
      selectedModels: models,
      onConfirm: setModels,
      onSetDirty,
    })
  }

  const handleSubmit = async () => {
    setSubmitting(true)
    try {
      await approvalApi.create({
        type,
        reason,
        requestedQuota: Number(quota),
        requestedModels: models,
        memberId,
      })
      toast.success('申请已提交')
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('提交失败')
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
          primaryDisabled={!reason.trim() || submitting}
        />
      }
    >
      <div className="max-w-lg space-y-5">
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
        <div className="space-y-3">
          <Label>申请模型</Label>
          <Button variant="outline" onClick={handlePickModels}>
            选择模型 ({models.length})
          </Button>
          <div className="flex flex-wrap gap-1">
            {models.map((m) => (
              <Badge key={m} variant="outline" className="text-xs">
                {m}
              </Badge>
            ))}
          </div>
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}

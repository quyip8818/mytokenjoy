import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { isCustomModel } from '@/features/models'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'

export function ModelEditWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'model-edit'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const model = entry.payload.model
  const onSuccess = entry.payload.onSuccess as ((id?: string | number) => void) | undefined
  const [label, setLabel] = useState(model.name)
  const [description, setDescription] = useState(model.description)
  const [endpoint, setEndpoint] = useState(model.endpoint ?? '')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!label.trim()) return
    setSubmitting(true)
    try {
      await apis.modelApi.update(model.modelId, {
        name: label.trim(),
        description: description.trim(),
        endpoint: isCustomModel(model) ? endpoint.trim() : undefined,
      })
      toast.success('模型信息已更新')
      onSuccess?.(model.modelId)
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '更新失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="编辑自定义模型"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中...' : '保存'}
          onPrimary={handleSubmit}
          primaryDisabled={!label.trim() || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        <div className="space-y-1.5">
          <Label>显示名称</Label>
          <Input
            value={label}
            onChange={(e) => {
              setLabel(e.target.value)
              onSetDirty(true)
            }}
          />
        </div>
        <div className="space-y-1.5">
          <Label>描述</Label>
          <Textarea
            value={description}
            onChange={(e) => {
              setDescription(e.target.value)
              onSetDirty(true)
            }}
            rows={3}
          />
        </div>
        {isCustomModel(model) && (
          <div className="space-y-1.5">
            <Label>部署地址</Label>
            <Input
              value={endpoint}
              onChange={(e) => {
                setEndpoint(e.target.value)
                onSetDirty(true)
              }}
              placeholder="https://api.example.com/v1"
            />
          </div>
        )}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import type { ModelVisibility } from '@/api/types'
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
import { useWorkflow } from '../use-workflow'
import { workflowErrorMessage } from '../lib/error-message'

const VISIBILITY_OPTIONS: { value: ModelVisibility; label: string }[] = [
  { value: 'all', label: '全员可见' },
  { value: 'department', label: '部门可见' },
  { value: 'custom', label: '自定义' },
]

export function ModelEditWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'model-edit'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const model = entry.payload.model
  const onSuccess = entry.payload.onSuccess as ((id?: string) => void) | undefined
  const [displayName, setDisplayName] = useState(model.displayName)
  const [description, setDescription] = useState(model.description)
  const [visibility, setVisibility] = useState<ModelVisibility>(model.visibility)
  const [endpoint, setEndpoint] = useState(model.endpoint ?? '')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!displayName.trim()) return
    setSubmitting(true)
    try {
      await apis.modelApi.update(model.id, {
        displayName: displayName.trim(),
        description: description.trim(),
        visibility,
        endpoint: model.type === 'custom' ? endpoint.trim() : undefined,
      })
      toast.success('模型信息已更新')
      onSuccess?.(model.id)
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
          primaryDisabled={!displayName.trim() || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        <div className="space-y-1.5">
          <Label>显示名称</Label>
          <Input
            value={displayName}
            onChange={(e) => {
              setDisplayName(e.target.value)
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
        <div className="space-y-1.5">
          <Label>可见范围</Label>
          <Select
            value={visibility}
            onValueChange={(value) => {
              setVisibility(value as ModelVisibility)
              onSetDirty(true)
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {VISIBILITY_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">仅展示；运行时权限以 Key / 部门白名单为准</p>
        </div>
        {model.type === 'custom' && (
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
